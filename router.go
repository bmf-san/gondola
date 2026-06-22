package gondola

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"
)

// swappableHandler is an http.Handler whose delegate can be replaced atomically
// at runtime. It is used to hot-swap the router when the configuration is
// reloaded without dropping in-flight connections.
type swappableHandler struct {
	h atomic.Pointer[http.Handler]
}

func newSwappableHandler(h http.Handler) *swappableHandler {
	s := &swappableHandler{}
	s.h.Store(&h)
	return s
}

func (s *swappableHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	(*s.h.Load()).ServeHTTP(w, r)
}

func (s *swappableHandler) swap(h http.Handler) {
	s.h.Store(&h)
}

// newRouter builds the HTTP handler that serves static files and proxies
// requests to upstreams. Reverse proxies are constructed once here rather than
// per request.
func newRouter(c *Config, logger *slog.Logger) (http.Handler, error) {
	if err := validateUpstreams(c.Upstreams); err != nil {
		return nil, err
	}

	proxies := make(map[string]http.Handler, len(c.Upstreams))
	for _, up := range c.Upstreams {
		target, err := url.Parse(up.Target)
		if err != nil {
			return nil, fmt.Errorf("invalid upstream target URL %q: %w", up.Target, err)
		}

		rp := httputil.NewSingleHostReverseProxy(target)
		rp.Transport = NewLogRoundTripper(newUpstreamTransport(up))
		rp.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			logger.ErrorContext(r.Context(), "upstream request failed",
				slog.String("upstream", target.Host),
				slog.String("error", err.Error()),
			)
			w.WriteHeader(http.StatusBadGateway)
		}
		proxies[up.HostName] = NewProxyHandler(rp, logger)
	}

	staticFiles := c.Proxy.StaticFiles

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Static files take precedence over proxying.
		if serveStaticFile(w, r, staticFiles) {
			return
		}
		// Route to an upstream by Host header.
		if h, ok := proxies[r.Host]; ok {
			h.ServeHTTP(w, r)
			return
		}
		http.NotFound(w, r)
	})

	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		for _, sf := range staticFiles {
			p := filepath.Join(sf.Dir, "favicon.ico")
			if _, err := os.Stat(p); err == nil {
				http.ServeFile(w, r, p)
				return
			}
		}
		w.WriteHeader(http.StatusNoContent)
	})

	return mux, nil
}

// serveStaticFile attempts to serve the request from one of the configured
// static file rules. It returns true when the request was handled.
func serveStaticFile(w http.ResponseWriter, r *http.Request, staticFiles []StaticFile) bool {
	for _, sf := range staticFiles {
		if !strings.HasPrefix(r.URL.Path, sf.Path) {
			continue
		}

		p := strings.TrimPrefix(r.URL.Path, sf.Path)
		// Rewrite the request path so http.ServeFile's built-in protection
		// against "../" traversal applies to the trimmed path.
		r2 := r.Clone(r.Context())
		r2.URL.Path = p

		fullPath := filepath.Join(sf.Dir, p)
		fileInfo, err := os.Stat(fullPath)

		useFallback := false
		switch {
		case err != nil:
			useFallback = true
		case fileInfo.IsDir():
			indexPath := filepath.Join(fullPath, "index.html")
			if _, err := os.Stat(indexPath); err == nil {
				http.ServeFile(w, r2, indexPath)
				return true
			}
			useFallback = true
		}

		if useFallback {
			fallbackFile := "index.html"
			if sf.FallbackPath != "" {
				fallbackFile = sf.FallbackPath
			}
			http.ServeFile(w, r2, filepath.Join(sf.Dir, fallbackFile))
			return true
		}

		http.ServeFile(w, r2, fullPath)
		return true
	}
	return false
}

// validateTarget reports whether target is a usable absolute proxy URL.
func validateTarget(target string) error {
	u, err := url.Parse(target)
	if err != nil {
		return err
	}
	if u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("missing scheme or host")
	}
	return nil
}

// newUpstreamTransport builds an http.Transport honoring the upstream's
// per-target timeouts. ReadTimeout bounds the wait for response headers and
// WriteTimeout bounds connection establishment.
func newUpstreamTransport(up Upstream) http.RoundTripper {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   msToDuration(up.WriteTimeout),
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: msToDuration(up.ReadTimeout),
	}
}
