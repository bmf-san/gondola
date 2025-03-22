package gondola

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"testing"
)

type mockTransport struct {
	response *http.Response
	err      error
}

func (t *mockTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return t.response, t.err
}

func TestNewLogRoundTripper(t *testing.T) {
	transport := http.DefaultTransport
	lrt := NewLogRoundTripper(transport)
	if lrt == nil {
		t.Error("Expected LogRoundTripper to not be nil")
	}
	if lrt != nil && lrt.transport != transport {
		t.Errorf("Expected transport to be %v, got %v", transport, lrt.transport)
	}
}

func TestRoundTrip(t *testing.T) {
	tests := []struct {
		name          string
		transport     http.RoundTripper
		expectedError bool
	}{
		{
			name: "successful request",
			transport: &mockTransport{
				response: &http.Response{
					Status:        "200 OK",
					StatusCode:    http.StatusOK,
					Body:          io.NopCloser(bytes.NewBufferString("test")),
					ContentLength: 4,
				},
			},
			expectedError: false,
		},
		{
			name: "transport error",
			transport: &mockTransport{
				err: errors.New("mock transport error"),
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := slog.New(slog.NewJSONHandler(&buf, nil))
			slog.SetDefault(logger)

			lrt := NewLogRoundTripper(tt.transport)
			dummy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("dummy"))
			}))
			defer dummy.Close()

			req := httptest.NewRequest(http.MethodGet, dummy.URL, nil)
			req = SetInfo(req, &responseInfo{})
			resp, err := lrt.RoundTrip(req)

			if tt.expectedError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectedError {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if resp == nil {
					t.Error("Expected response but got nil")
				}
			}

			if !tt.expectedError {
				info := GetInfo(req)
				if info.upstreamStatus != "200 OK" {
					t.Errorf("Expected upstream status '200 OK', got '%s'", info.upstreamStatus)
				}
				if info.upstreamSize != 4 {
					t.Errorf("Expected upstream size 4, got %d", info.upstreamSize)
				}
				if info.upstreamTime <= 0 {
					t.Error("Expected upstream time > 0")
				}
			}
		})
	}
}

func TestProxyHandler(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		setupBackend   func() (*httptest.Server, error)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful proxy with root path",
			path: "/",
			setupBackend: func() (*httptest.Server, error) {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("success"))
				})), nil
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
		},
		{
			name: "successful proxy with sub path",
			path: "/foo",
			setupBackend: func() (*httptest.Server, error) {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("success"))
				})), nil
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backend, err := tt.setupBackend()
			if err != nil {
				t.Fatalf("Failed to setup backend: %v", err)
			}
			defer backend.Close()

			backendURL, err := url.Parse(backend.URL)
			if err != nil {
				t.Fatalf("Failed to parse backend URL: %v", err)
			}

			var buf bytes.Buffer
			logger := slog.New(slog.NewJSONHandler(&buf, nil))
			slog.SetDefault(logger)

			proxy := httputil.NewSingleHostReverseProxy(backendURL)
			proxy.Transport = NewLogRoundTripper(http.DefaultTransport)
			handler := NewProxyHandler(proxy, slog.New(slog.NewJSONHandler(&buf, nil)))

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("http://example.com%s", tt.path), nil)
			req.RemoteAddr = "192.0.2.1:12345" // Testing IP and port
			req = SetInfo(req, &responseInfo{})
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			resp := w.Result()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Failed to read response body: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
			if string(body) != tt.expectedBody {
				t.Errorf("Expected body %q, got %q", tt.expectedBody, string(body))
			}

			logOutput := buf.String()
			// Check basic log structure
			if !strings.Contains(logOutput, `"level":"INFO"`) {
				t.Error("Expected INFO log level")
			}
			if !strings.Contains(logOutput, `"msg":"access_log"`) {
				t.Error("Expected access_log message")
			}

			// Check client info
			if !strings.Contains(logOutput, `"remote_addr":"192.0.2.1"`) {
				t.Error("Expected remote_addr 192.0.2.1")
			}
			if !strings.Contains(logOutput, `"remote_port":"12345"`) {
				t.Error("Expected remote_port 12345")
			}
			if !strings.Contains(logOutput, `"x_forwarded_for":""`) {
				t.Error("Expected empty x_forwarded_for")
			}

			// Check request info
			if !strings.Contains(logOutput, `"method":"GET"`) {
				t.Error("Expected method GET")
			}
			fullURI := fmt.Sprintf("http://example.com%s", tt.path)
			if !strings.Contains(logOutput, fmt.Sprintf(`"request_uri":"%s"`, fullURI)) {
				t.Errorf("Expected request_uri %q", fullURI)
			}
			if !strings.Contains(logOutput, `"query_string":""`) {
				t.Error("Expected empty query_string")
			}
			if !strings.Contains(logOutput, `"host":"example.com"`) {
				t.Error("Expected host example.com")
			}
			if !strings.Contains(logOutput, `"request_size":0`) {
				t.Error("Expected request_size 0")
			}

			// Check response info
			if !strings.Contains(logOutput, `"status":"OK"`) {
				t.Error("Expected status OK")
			}
			if !strings.Contains(logOutput, `"body_bytes_sent":7`) {
				t.Error("Expected body_bytes_sent 7")
			}
			if !strings.Contains(logOutput, `"bytes_sent":7`) {
				t.Error("Expected bytes_sent 7")
			}
			if !strings.Contains(logOutput, `"request_time":`) {
				t.Error("Expected request_time field")
			}

			// Check upstream info
			if !strings.Contains(logOutput, `"upstream_addr":`) {
				t.Error("Expected upstream_addr field")
			}
			if !strings.Contains(logOutput, `"upstream_status":"200 OK"`) {
				t.Error("Expected upstream_status 200 OK")
			}
			if !strings.Contains(logOutput, `"upstream_size":7`) {
				t.Error("Expected upstream_size 7")
			}
			if !strings.Contains(logOutput, `"upstream_response_time":`) {
				t.Error("Expected upstream_response_time field")
			}

			// Check header info
			if !strings.Contains(logOutput, `"referer":""`) {
				t.Error("Expected empty referer")
			}
			if !strings.Contains(logOutput, `"user_agent":""`) {
				t.Error("Expected empty user_agent")
			}
		})
	}
}

func TestNewServer(t *testing.T) {
	tests := []struct {
		name          string
		config        string
		expectedError bool
	}{
		{
			name: "valid configuration",
			config: `
proxy:
  port: "8080"
  read_header_timeout: 2000
  shutdown_timeout: 3000
  static_files:
    - path: /public/
      dir: testdata/public
upstreams:
  - host_name: backend1.local
    target: http://localhost:8081
`,
			expectedError: false,
		},
		{
			name: "invalid target URL",
			config: `
proxy:
  port: "8080"
upstreams:
  - host_name: backend1.local
    target: :invalid:url
`,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{}
			c, err := cfg.Load(strings.NewReader(tt.config))
			if err != nil {
				t.Fatalf("Failed to load config: %v", err)
			}

			server, err := NewServer(c)
			if tt.expectedError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if server == nil {
					t.Error("Expected server but got nil")
				} else {
					server.Close()
				}
			}
		})
	}
}

func TestStaticFileHandler(t *testing.T) {
	content := []byte("test content")
	tmpDir := t.TempDir()
	if err := os.WriteFile(tmpDir+"/test.txt", content, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	cfg := &Config{
		Proxy: Proxy{
			StaticFiles: []StaticFile{
				{
					Path: "/static/",
					Dir:  tmpDir,
				},
			},
		},
	}

	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	defer server.Close()

	ts := httptest.NewServer(server.Handler)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/static/test.txt")
	if err != nil {
		t.Fatalf("Failed to get static file: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
	if string(body) != string(content) {
		t.Errorf("Expected body %q, got %q", string(content), string(body))
	}
}
