package gondola

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
)

// Runner is an interface that defines the Run method.
type Runner interface {
	Run() error
}

// Gondola is a reverse proxy server.
type Gondola struct {
	config     *Config
	logger     *Logger
	server     *http.Server
	router     *swappableHandler
	configPath string
	logLevel   string // optional override (e.g. from GONDOLA_LOG_LEVEL)
}

// Option configures a Gondola instance.
type Option func(*Gondola)

// WithConfigPath enables hot-reloading the configuration file on SIGHUP by
// telling Gondola where to re-read it from.
func WithConfigPath(path string) Option {
	return func(g *Gondola) { g.configPath = path }
}

// WithLogLevel overrides the log level from the configuration file (debug,
// info, warn, error).
func WithLogLevel(level string) Option {
	return func(g *Gondola) { g.logLevel = level }
}

// NewGondola returns a new Gondola built from the configuration in r.
func NewGondola(r io.Reader, opts ...Option) (*Gondola, error) {
	cfg := &Config{}
	c, err := cfg.Load(r)
	if err != nil {
		return nil, &ConfigLoadError{Err: err}
	}
	c.setDefaults()

	g := &Gondola{config: c}
	for _, opt := range opts {
		opt(g)
	}

	level := c.SlogLevel()
	if g.logLevel != "" {
		level = ParseLevel(g.logLevel)
	}

	logger, err := NewLogger(level, c.Proxy.LogFile, c.LogFormat, c.LogCustomFormat)
	if err != nil {
		return nil, &ProxyServerError{Err: err}
	}
	g.logger = logger

	router, err := newRouter(c, logger.Logger)
	if err != nil {
		return nil, &ProxyServerError{Err: err}
	}
	g.router = newSwappableHandler(router)
	g.server = newHTTPServer(c, g.router)

	return g, nil
}

// NewServer creates a new configured HTTP server for the given configuration.
// It is primarily useful for testing and embedding; most callers should use
// NewGondola and Run.
func NewServer(c *Config) (*http.Server, error) {
	c.setDefaults()

	logger, err := NewLogger(c.SlogLevel(), c.Proxy.LogFile, c.LogFormat, c.LogCustomFormat)
	if err != nil {
		return nil, err
	}

	router, err := newRouter(c, logger.Logger)
	if err != nil {
		return nil, err
	}

	return newHTTPServer(c, newSwappableHandler(router)), nil
}

// newHTTPServer builds an *http.Server with timeouts and TLS settings derived
// from the configuration.
func newHTTPServer(c *Config, handler http.Handler) *http.Server {
	s := &http.Server{
		Addr:              ":" + c.Proxy.Port,
		Handler:           handler,
		ReadHeaderTimeout: msToDuration(c.Proxy.ReadHeaderTimeout),
		ReadTimeout:       msToDuration(c.Proxy.ReadTimeout),
		WriteTimeout:      msToDuration(c.Proxy.WriteTimeout),
		IdleTimeout:       msToDuration(c.Proxy.IdleTimeout),
		MaxHeaderBytes:    c.Proxy.MaxHeaderBytes,
	}
	if c.Proxy.IsEnableTLS() {
		s.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	}
	return s
}

// Run starts the proxy server and blocks until it is shut down by a signal or
// stops with an error. It handles:
//
//	SIGINT, SIGTERM -> graceful shutdown
//	SIGHUP          -> reload configuration (when a config path is known)
//	SIGUSR1         -> reopen access logs (for log rotation)
func (g *Gondola) Run() error {
	slog.SetDefault(g.logger.Logger)

	errCh := make(chan error, 1)
	go func() {
		errCh <- g.serve()
	}()

	sigCh := make(chan os.Signal, 1)
	notifySignals(sigCh)
	defer signal.Stop(sigCh)

	for {
		select {
		case err := <-errCh:
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				return fmt.Errorf("server stopped: %w", err)
			}
			return nil
		case sig := <-sigCh:
			if g.handleSignal(sig) {
				return g.Shutdown()
			}
		}
	}
}

// serve starts listening, with or without TLS depending on the configuration.
func (g *Gondola) serve() error {
	if g.config.Proxy.IsEnableTLS() {
		g.logger.Info("starting server with TLS", slog.String("port", g.config.Proxy.Port))
		return g.server.ListenAndServeTLS(g.config.Proxy.TLSCertPath, g.config.Proxy.TLSKeyPath)
	}
	g.logger.Info("starting server", slog.String("port", g.config.Proxy.Port))
	return g.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server, waiting up to the configured
// shutdown timeout for in-flight requests to complete.
func (g *Gondola) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), msToDuration(g.config.Proxy.ShutdownTimeout))
	defer cancel()

	g.logger.Info("graceful shutdown started")
	err := g.server.Shutdown(ctx)
	if err != nil {
		g.logger.Error("graceful shutdown failed", slog.String("error", err.Error()))
	} else {
		g.logger.Info("graceful shutdown completed")
	}
	_ = g.logger.Close()

	if err != nil {
		return fmt.Errorf("graceful shutdown failed: %w", err)
	}
	return nil
}

// reload re-reads the configuration file and atomically swaps the router.
// Changes to the listening port and server-level timeouts require a restart
// and are not applied by reload.
func (g *Gondola) reload() error {
	if g.configPath == "" {
		g.logger.Warn("no config path configured; skipping reload")
		return nil
	}

	// #nosec G304 -- path is operator-provided configuration, not user input.
	f, err := os.Open(filepath.Clean(g.configPath))
	if err != nil {
		return fmt.Errorf("open config: %w", err)
	}
	defer func() { _ = f.Close() }()

	cfg := &Config{}
	c, err := cfg.Load(f)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	c.setDefaults()

	router, err := newRouter(c, g.logger.Logger)
	if err != nil {
		return fmt.Errorf("build router: %w", err)
	}

	g.router.swap(router)
	g.config = c
	g.logger.Info("configuration reloaded")
	return nil
}
