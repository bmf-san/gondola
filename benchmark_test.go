package gondola

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

// benchDiscardLogger returns a logger that discards output so logging does not
// skew benchmark measurements.
func benchDiscardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// BenchmarkProxyHandler measures reverse-proxy request handling for a single
// matched upstream.
func BenchmarkProxyHandler(b *testing.B) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	defer backend.Close()

	cfg := &Config{Upstreams: []Upstream{{HostName: "backend.local", Target: backend.URL}}}
	cfg.setDefaults()
	handler, err := newRouter(cfg, benchDiscardLogger())
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "http://backend.local/", nil)
		req.Host = "backend.local"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}

// BenchmarkProxyHandlerParallel measures concurrent reverse-proxy handling.
func BenchmarkProxyHandlerParallel(b *testing.B) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	defer backend.Close()

	cfg := &Config{Upstreams: []Upstream{{HostName: "backend.local", Target: backend.URL}}}
	cfg.setDefaults()
	handler, err := newRouter(cfg, benchDiscardLogger())
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest(http.MethodGet, "http://backend.local/", nil)
			req.Host = "backend.local"
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
		}
	})
}

// BenchmarkStaticFile measures static file serving.
func BenchmarkStaticFile(b *testing.B) {
	cfg := &Config{Proxy: Proxy{StaticFiles: []StaticFile{{Path: "/static/", Dir: "testdata/static"}}}}
	cfg.setDefaults()
	handler, err := newRouter(cfg, benchDiscardLogger())
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "http://localhost/static/test.txt", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}

// BenchmarkRouterBuild measures the cost of building the router (reverse
// proxies are constructed once here at startup / on reload).
func BenchmarkRouterBuild(b *testing.B) {
	cfg := &Config{
		Upstreams: []Upstream{
			{HostName: "a.local", Target: "http://127.0.0.1:8081"},
			{HostName: "b.local", Target: "http://127.0.0.1:8082"},
		},
	}
	cfg.setDefaults()
	logger := benchDiscardLogger()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := newRouter(cfg, logger); err != nil {
			b.Fatal(err)
		}
	}
}
