package gondola

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Runner is an interface that defines the Run method.
type Runner interface {
	Run() error
}

// NewGondola returns a new Gondola.
func NewGondola(r io.Reader) (*Gondola, error) {
	cfg := &Config{}
	c, err := cfg.Load(r)
	if err != nil {
		return nil, &ConfigLoadError{Err: err}
	}

	s, err := NewServer(c)
	if err != nil {
		return nil, &ProxyServerError{Err: err}
	}

	return &Gondola{
		config: c,
		server: s,
	}, nil
}

// NewServer creates a new HTTP server with the given configuration.
func NewServer(c *Config) (*http.Server, error) {
	mux := http.NewServeMux()
	logger := NewLogger(c.LogLevel)

	// Validate upstream configurations first
	for _, upstream := range c.Upstreams {
		if _, err := url.Parse(upstream.Target); err != nil {
			return nil, fmt.Errorf("invalid upstream target URL %s: %w", upstream.Target, err)
		}
	}

	// Create a main handler that will handle both static files and proxy requests
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// First, try to serve static files
		for _, sf := range c.Proxy.StaticFiles {
			// Check if the request path starts with the configured path
			if strings.HasPrefix(r.URL.Path, sf.Path) {
				p := strings.TrimPrefix(r.URL.Path, sf.Path)
				r2 := new(http.Request)
				*r2 = *r
				r2.URL = new(url.URL)
				*r2.URL = *r.URL
				r2.URL.Path = p

				fullPath := filepath.Join(sf.Dir, p)
				fileInfo, err := os.Stat(fullPath)

				// Always use fallback for directory requests or non-existent files
				var useFallback bool
				if err != nil {
					useFallback = true
				} else if fileInfo.IsDir() {
					// If directory and has index.html, serve it
					indexPath := filepath.Join(fullPath, "index.html")
					if _, err := os.Stat(indexPath); err == nil {
						http.ServeFile(w, r2, indexPath)
						return
					}
					useFallback = true
				}

				if useFallback {
					fallbackFile := "index.html"
					if sf.FallbackPath != "" {
						fallbackFile = sf.FallbackPath
					}
					http.ServeFile(w, r2, filepath.Join(sf.Dir, fallbackFile))
					return
				}

				// Serve the existing file
				http.ServeFile(w, r2, fullPath)
				return
			}
		}

		// If no static file is matched, try to proxy the request
		for _, upstream := range c.Upstreams {
			if r.Host == upstream.HostName {
				target, err := url.Parse(upstream.Target)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				proxy := httputil.NewSingleHostReverseProxy(target)
				proxy.Transport = NewLogRoundTripper(http.DefaultTransport)
				proxy.Director = func(req *http.Request) {
					req.URL.Scheme = target.Scheme
					req.URL.Host = target.Host
				}
				handler := NewProxyHandler(proxy, logger.Logger)
				handler.ServeHTTP(w, r)
				return
			}
		}
	})

	// Handle favicon.ico requests
	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		for _, sf := range c.Proxy.StaticFiles {
			if _, err := os.Stat(filepath.Join(sf.Dir, "favicon.ico")); err == nil {
				http.ServeFile(w, r, filepath.Join(sf.Dir, "favicon.ico"))
				return
			}
		}
		w.WriteHeader(http.StatusNoContent)
	})

	server := &http.Server{
		Addr:              ":" + c.Proxy.Port,
		ReadHeaderTimeout: time.Duration(c.Proxy.ReadHeaderTimeout) * time.Millisecond,
		Handler:           mux,
	}

	return server, nil
}

// TODO: Need to dynamically load a configuration file. For now, we will limit the implementation to just loading the file at startup.
// Run starts the proxy server.
func (g *Gondola) Run() error {
	logger := NewLogger(g.config.LogLevel)
	slog.SetDefault(logger.Logger)

	// TODO: do health check for upstreams.

	if g.config.Proxy.IsEnableTLS() {
		slog.Info(fmt.Sprintf("Running server on port %s with TLS...", g.config.Proxy.Port))
		if err := g.server.ListenAndServeTLS(g.config.Proxy.TLSCertPath, g.config.Proxy.TLSKeyPath); err != nil {
			slog.Error("Error running server with TLS: " + err.Error())
		}
	} else {
		slog.Info("Running server on port " + g.config.Proxy.Port + "...")
		if err := g.server.ListenAndServe(); err != nil {
			slog.Error("Error running server: " + err.Error())
		}
	}
	return nil
}
