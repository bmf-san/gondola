package gondola

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseErrorPages(t *testing.T) {
	pages, def := parseErrorPages(map[string]string{
		"404":             "/404.html",
		"500 502 503 504": "/50x.html",
		"default":         "/error.html",
	}, "testdata/static")

	if got, want := pages[404], filepath.Join("testdata/static", "404.html"); got != want {
		t.Errorf("404 -> %q, want %q", got, want)
	}
	for _, c := range []int{500, 502, 503, 504} {
		if got, want := pages[c], filepath.Join("testdata/static", "50x.html"); got != want {
			t.Errorf("%d -> %q, want %q", c, got, want)
		}
	}
	if got, want := def, filepath.Join("testdata/static", "error.html"); got != want {
		t.Errorf("default -> %q, want %q", got, want)
	}

	if p, d := parseErrorPages(nil, ""); p != nil || d != "" {
		t.Errorf("empty config should yield nil/empty, got %v/%q", p, d)
	}
}

func TestErrorPagesMiddlewareIntercepts(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	})
	mw := errorPagesMiddleware(handler, map[int]string{500: "testdata/static/404.html"}, "")

	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500 preserved, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "custom 404 content") {
		t.Errorf("expected custom error page body, got %q", rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "boom") {
		t.Errorf("original body should be suppressed, got %q", rec.Body.String())
	}
}

func TestErrorPagesMiddlewarePassthrough(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mw := errorPagesMiddleware(handler, map[int]string{404: "testdata/static/404.html"}, "")

	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK || rec.Body.String() != "ok" {
		t.Errorf("passthrough failed: code=%d body=%q", rec.Code, rec.Body.String())
	}
}

func TestErrorPagesMiddlewareMissingFile(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	})
	mw := errorPagesMiddleware(handler, map[int]string{500: "testdata/static/does_not_exist.html"}, "")

	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "boom") {
		t.Errorf("expected fallback to built-in body, got %q", rec.Body.String())
	}
}

func TestNewRouterErrorPages(t *testing.T) {
	cfg := &Config{Proxy: Proxy{
		StaticFiles: []StaticFile{{Path: "/static/", Dir: "testdata/static"}},
		ErrorPages:  map[string]string{"404": "/404.html"},
	}}
	cfg.setDefaults()

	h, err := newRouter(cfg, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "http://x/nope", nil))
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "custom 404 content") {
		t.Errorf("expected custom 404 page, got %q", rec.Body.String())
	}
}
