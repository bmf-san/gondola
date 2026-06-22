package gondola

import (
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// parseErrorPages converts the raw error_pages configuration into a lookup of
// HTTP status code -> resolved file path, plus an optional default page. Page
// paths are resolved relative to root (the first static file directory).
//
// Keys may be a single code ("404"), a space-separated group ("500 502 503
// 504"), or "default".
func parseErrorPages(raw map[string]string, root string) (map[int]string, string) {
	if len(raw) == 0 {
		return nil, ""
	}
	pages := make(map[int]string)
	var defaultPage string
	for key, path := range raw {
		resolved := filepath.Join(root, strings.TrimPrefix(path, "/"))
		if strings.EqualFold(strings.TrimSpace(key), "default") {
			defaultPage = resolved
			continue
		}
		for _, field := range strings.Fields(key) {
			if code, err := strconv.Atoi(field); err == nil {
				pages[code] = resolved
			}
		}
	}
	return pages, defaultPage
}

// errorPagesMiddleware serves configured custom error pages for error status
// codes (>= 400). When no pages are configured it returns next unchanged.
func errorPagesMiddleware(next http.Handler, pages map[int]string, defaultPage string) http.Handler {
	if len(pages) == 0 && defaultPage == "" {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(&errorPageWriter{
			ResponseWriter: w,
			pages:          pages,
			defaultPage:    defaultPage,
		}, r)
	})
}

// errorPageWriter intercepts error responses and replaces their body with the
// configured custom error page while preserving the original status code.
type errorPageWriter struct {
	http.ResponseWriter
	pages       map[int]string
	defaultPage string
	handled     bool
	intercepted bool
}

func (w *errorPageWriter) page(code int) (string, bool) {
	if p, ok := w.pages[code]; ok {
		return p, true
	}
	if w.defaultPage != "" {
		return w.defaultPage, true
	}
	return "", false
}

func (w *errorPageWriter) WriteHeader(code int) {
	if w.handled {
		return
	}
	w.handled = true

	if code >= 400 {
		if page, ok := w.page(code); ok && w.servePage(code, page) {
			w.intercepted = true
			return
		}
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *errorPageWriter) Write(b []byte) (int, error) {
	if w.intercepted {
		// Suppress the original (upstream) body; the error page was written.
		return len(b), nil
	}
	if !w.handled {
		w.WriteHeader(http.StatusOK)
		if w.intercepted {
			return len(b), nil
		}
	}
	return w.ResponseWriter.Write(b)
}

// servePage writes the custom error page with the given status code. It returns
// false (so the caller falls back to the built-in response) when the page file
// cannot be read.
func (w *errorPageWriter) servePage(code int, page string) bool {
	// #nosec G304 -- page path is operator-provided configuration, not user input.
	content, err := os.ReadFile(page)
	if err != nil {
		return false
	}

	h := w.ResponseWriter.Header()
	h.Del("Content-Length")
	h.Set("Content-Type", contentTypeFor(page))
	h.Set("Content-Length", strconv.Itoa(len(content)))
	w.ResponseWriter.WriteHeader(code)
	_, _ = w.ResponseWriter.Write(content)
	return true
}

// contentTypeFor returns the MIME type for an error page based on its
// extension, defaulting to HTML.
func contentTypeFor(path string) string {
	if ct := mime.TypeByExtension(filepath.Ext(path)); ct != "" {
		return ct
	}
	return "text/html; charset=utf-8"
}
