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

	// Set up static file handlers
	for _, sf := range c.Proxy.StaticFiles {
		fs := http.FileServer(http.Dir(sf.Dir))
		mux.Handle(sf.Path, http.StripPrefix(sf.Path, fs))
	}

	// Handle favicon.ico requests
	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		// Check if favicon exists in any of the static file directories
		for _, sf := range c.Proxy.StaticFiles {
			if _, err := os.Stat(filepath.Join(sf.Dir, "favicon.ico")); err == nil {
				http.ServeFile(w, r, filepath.Join(sf.Dir, "favicon.ico"))
				return
			}
		}
		// If no favicon found, return 204
		w.WriteHeader(http.StatusNoContent)
	})

	// Set up proxy handlers for each upstream
	for _, upstream := range c.Upstreams {
		target, err := url.Parse(upstream.Target)
		if err != nil {
			return nil, fmt.Errorf("invalid upstream target URL %s: %w", upstream.Target, err)
		}

		proxy := httputil.NewSingleHostReverseProxy(target)
		proxy.Transport = NewLogRoundTripper(http.DefaultTransport)
		proxy.Director = func(req *http.Request) {
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
		}
		handler := NewProxyHandler(proxy, logger.Logger)
		pattern := upstream.HostName + "/"
		mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
			if r.Host == upstream.HostName {
				handler.ServeHTTP(w, r)
			}
		})
	}

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
