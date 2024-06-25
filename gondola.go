package gondola

import (
	"fmt"
	"io"
	"log/slog"
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

	s, err := newServer(c)
	if err != nil {
		return nil, &ProxyServerError{Err: err}
	}

	return &Gondola{
		config: c,
		server: s,
	}, nil
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
