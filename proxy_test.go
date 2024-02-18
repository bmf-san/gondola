package gondola

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
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

	r := strings.NewReader(data)
	gondola, err := NewGondola(r)
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

func TestNewLogRoundTripper(t *testing.T) {
	transport := http.DefaultTransport
	lrt := NewLogRoundTripper(transport)
	if lrt == nil {
		t.Errorf("Expected LogRoundTripper, got nil")
	}
}

func TestRoundTrip(t *testing.T) {
	lrt := &LogRoundTripper{
		transport: http.DefaultTransport,
	}
	dummy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("dummy"))
	}))
	defer dummy.Close()
	dummyURL, err := url.Parse(dummy.URL)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	resp, err := lrt.RoundTrip(httptest.NewRequest(http.MethodGet, dummyURL.String(), nil))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp == nil {
		t.Errorf("Expected response, got nil")
	}
}

func TestHandler(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("backend"))
	}))
	defer backend.Close()

	backendURL, err := url.Parse(backend.URL)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxyHandler := &ProxyHandler{
			proxy: httputil.NewSingleHostReverseProxy(backendURL),
		}
		proxyHandler.Handler(w, r)
	}))
	defer proxy.Close()

	proxyURL, err := url.Parse(proxy.URL)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	resp, err := http.Get(proxyURL.String())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if string(b) != "backend" {
		t.Errorf("Expected body %s, got %s", "backend", string(b))
	}
}

func TestNewServer(t *testing.T) {
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

	cfg := &Config{}
	c, err := cfg.Load(strings.NewReader(data))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	server, err := newServer(c)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if server == nil {
		t.Errorf("Expected server, got nil")
	}
	server.Close()
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
