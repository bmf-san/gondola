package gondola

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"
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

func TestGracefulShutdown(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(2) // Wait for server startup and shutdown completion

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond) // Simulated processing time
		w.WriteHeader(http.StatusOK)
	})

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	server := NewProxyServer(handler, logger)

	// Channel to signal server readiness
	ready := make(chan struct{})

	// Start server in a separate goroutine
	go func() {
		defer wg.Done()
		server.mu.Lock()
		s := &http.Server{
			Addr:    ":0", // Let the system choose an available port
			Handler: server.handler,
		}
		server.server = s
		server.mu.Unlock()

		// Signal server readiness
		close(ready)

		if err := s.ListenAndServe(); err != http.ErrServerClosed {
			t.Errorf("unexpected server error: %v", err)
		}
	}()

	// Wait for server readiness
	<-ready

	// Simulate an ongoing request
	reqDone := make(chan struct{})
	go func() {
		defer close(reqDone)
		defer wg.Done()

		// Safely get server address
		server.mu.RLock()
		addr := server.server.Addr
		server.mu.RUnlock()

		_, err := http.Get("http://localhost" + addr)
		if err == nil {
			t.Error("expected error due to server shutdown")
		}
	}()

	// Initiate shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		t.Errorf("unexpected shutdown error: %v", err)
	}

	// Check shutdown flag
	if !server.IsShutdown() {
		t.Error("expected server to be marked as shutdown")
	}

	// Wait for all goroutines to complete
	wg.Wait()
}

func TestStaticFileHandler(t *testing.T) {
	tests := []struct {
		name         string
		requestPath  string
		fallbackPath string
		expectedCode int
		expectedBody string
	}{
		{
			name:         "existing file",
			requestPath:  "/static/test.txt",
			expectedCode: http.StatusOK,
			expectedBody: "test content\n",
		},
		{
			name:         "existing file in subdir",
			requestPath:  "/static/subdir/test.txt",
			expectedCode: http.StatusOK,
			expectedBody: "subdir content\n",
		},
		{
			name:         "non-existent file fallback to custom file",
			requestPath:  "/static/nonexistent.txt",
			fallbackPath: "404.html",
			expectedCode: http.StatusOK,
			expectedBody: "custom 404 content\n",
		},
		{
			name:         "directory request fallback to index.html",
			requestPath:  "/static/subdir/",
			expectedCode: http.StatusOK,
			expectedBody: "index content\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Proxy: Proxy{
					StaticFiles: []StaticFile{
						{
							Path:         "/static/",
							Dir:          "testdata/static",
							FallbackPath: tt.fallbackPath,
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

			resp, err := http.Get(ts.URL + tt.requestPath)
			if err != nil {
				t.Fatalf("Failed to get static file: %v", err)
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Failed to read response body: %v", err)
			}

			if resp.StatusCode != tt.expectedCode {
				t.Errorf("Expected status %d, got %d", tt.expectedCode, resp.StatusCode)
			}
			if string(body) != tt.expectedBody {
				t.Errorf("Expected body %q, got %q", tt.expectedBody, string(body))
			}
		})
	}
}
