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
