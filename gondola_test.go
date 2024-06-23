package gondola

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestNewGondola(t *testing.T) {
	data := `
proxy:
  port: 8080
  read_header_timeout: 2000
  shutdown_timeout: 3000
  tls_cert_path: /path/to/cert
  tls_key_path: /path/to/key
  static_files:
    - path: /public/
      dir: testdata/public
upstreams:
  - host_name: backend1.local
    target: http://backend1:8081
  - host_name: backend2.local
    target: http://backend2:8082
log_level: -4
`
	gondola, err := NewGondola(strings.NewReader(data))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if gondola.config == nil {
		t.Errorf("Expected config, got nil")
	}
	if gondola.server == nil {
		t.Errorf("Expected server, got nil")
	}
	gondola.server.Close()
}

func TestNewGondolaConfigLoadError(t *testing.T) {
	r := strings.NewReader("invalid")
	_, err := NewGondola(r)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
	var cfgErr *ConfigLoadError
	if !errors.As(err, &cfgErr) {
		t.Errorf("Expected error, got %v", err)
	}
}

func TestNewGondolaProxyServerError(t *testing.T) {
	data := `
proxy:
  port: 8080
  read_header_timeout: 2000
  shutdown_timeout: 3000
  tls_cert_path: /path/to/cert
  tls_key_path: /path/to/key
  static_files:
    - path: /public/
      dir: testdata/public
upstreams:
  - host_name: backend1.local
    target: "://"
  - host_name: backend2.local
    target: "://"
log_level: -4
`
	_, err := NewGondola(strings.NewReader(data))
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
	var psErr *ProxyServerError
	if !errors.As(err, &psErr) {
		t.Errorf("Expected error, got %v", err)
	}
}

func TestRun(t *testing.T) {
	backend1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("backend1"))
	}))
	defer backend1.Close()

	backend2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("backend2"))
	}))
	defer backend2.Close()

	backend1URL, err := url.Parse(backend1.URL)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	backend2URL, err := url.Parse(backend2.URL)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// without TLS config
	data := `
proxy:
  port: 8080
  read_header_timeout: 2000
  shutdown_timeout: 3000
  static_files:
    - path: /public/
      dir: testdata/public
upstreams:
  - host_name: backend1.local
    target: ` + backend1URL.String() + `
  - host_name: backend2.local
    target: ` + backend2URL.String() + `
log_level: -4
`
	gondola, err := NewGondola(strings.NewReader(data))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	go func() {
		gondola.Run()
	}()

	for _, test := range []struct {
		name    string
		reqPath string
		path    string
		body    string
		code    int
	}{
		{
			name:    "request to backend1",
			reqPath: "http://backend1.local:8080/",
			body:    "backend1",
			code:    http.StatusOK,
		},
		{
			name:    "request to backend2",
			reqPath: "http://backend2.local:8080/",
			body:    "backend2",
			code:    http.StatusOK,
		},
		{
			name:    "request to static file",
			reqPath: "http://localhost:8080/public/index.html",
			body:    "test",
			code:    http.StatusOK,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, test.reqPath, nil)
			if err != nil {
				t.Fatal(err)
			}
			res, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer res.Body.Close()
			b, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}
			if res.StatusCode != test.code {
				t.Errorf("Expected status code %d, got %d", test.code, res.StatusCode)
			}
			if string(b) != test.body {
				t.Errorf("Expected body %s, got %s", test.body, string(b))
			}
		})
	}
	gondola.server.Close()
}
