package gondola

import (
	"bytes"
	"errors"
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
			name:          "successful request",
			transport:     http.DefaultTransport,
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

			lrt := &LogRoundTripper{transport: tt.transport}
			dummy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("dummy"))
			}))
			defer dummy.Close()

			req := httptest.NewRequest(http.MethodGet, dummy.URL, nil)
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
				logOutput := buf.String()
				if !strings.Contains(logOutput, `"level":"INFO"`) {
					t.Error("Expected INFO log level")
				}
				if !strings.Contains(logOutput, `"msg":"upstream_response"`) {
					t.Error("Expected upstream_response message")
				}
			}
		})
	}
}

func TestProxyHandler(t *testing.T) {
	tests := []struct {
		name           string
		setupBackend   func() (*httptest.Server, error)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful proxy",
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

			handler := &ProxyHandler{
				proxy: httputil.NewSingleHostReverseProxy(backendURL),
			}

			req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
			w := httptest.NewRecorder()

			handler.Handler(w, req)

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
			if !strings.Contains(logOutput, `"level":"INFO"`) {
				t.Error("Expected INFO log level")
			}
			if !strings.Contains(logOutput, `"msg":"proxy_response"`) {
				t.Error("Expected proxy_response message")
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

			server, err := newServer(c)
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

	server, err := newServer(cfg)
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
