package gondola

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

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

	ch := make(chan error, 1)
	go func() {
		if g.config.Proxy.IsEnableTLS() {
			slog.Info(fmt.Sprintf("Running server on port %s with TLS...", g.config.Proxy.Port))
			if err := g.server.ListenAndServeTLS(g.config.Proxy.TLSCertPath, g.config.Proxy.TLSKeyPath); err != http.ErrServerClosed {
				ch <- err
			}
		} else {
			slog.Info("Running server on port " + g.config.Proxy.Port + "...")
			if err := g.server.ListenAndServe(); err != http.ErrServerClosed {
				ch <- err
			}
		}
	}()
	e := <-ch
	if e != nil {
		return e
	}

	q := make(chan os.Signal, 1)
	signal.Notify(q, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer signal.Stop(q)
	<-q

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(g.config.Proxy.ShutdownTimeout)*time.Millisecond)
	defer cancel()
	if err := g.server.Shutdown(ctx); err != nil {
		slog.Error(fmt.Sprintf("Shutdown failed: %v", err))
		return err
	}

	slog.Info("Server stopped gracefully")
	return nil
}
