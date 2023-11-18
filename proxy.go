package gondola

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// CLI is a command line interface.
type CLI interface {
	Run()
}

// Gondola is a proxy server.
type Gondola struct {
	logger *slog.Logger
	config *Config
	server *http.Server
}

// NewGondola returns a new Gondola.
func NewGondola(l *slog.Logger, r io.Reader) (*Gondola, error) {
	c, err := loadConfig(r)
	if err != nil {
		return nil, err
	}

	s, err := newServer(c)
	if err != nil {
		return nil, err
	}

	return &Gondola{
		logger: l,
		config: c,
		server: s,
	}, nil
}

// loadConfig loads a configuration file.
func loadConfig(r io.Reader) (*Config, error) {
	cfg := &Config{}
	c, err := cfg.Load(r)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// newServer returns a new http.Server.
func newServer(c *Config) (*http.Server, error) {
	s := &http.Server{
		Addr:              ":" + c.Proxy.Port,
		ReadHeaderTimeout: time.Duration(c.Proxy.ReadHeaderTimeout) * time.Millisecond,
	}

	for _, b := range c.Upstreams {
		pp, err := url.Parse(b.Target)
		if err != nil {
			return nil, fmt.Errorf("error parsing upstream address: %w", err)
		}

		proxy := httputil.NewSingleHostReverseProxy(pp)
		http.HandleFunc(b.HostName+"/", func(w http.ResponseWriter, r *http.Request) {
			proxy.ServeHTTP(w, r)
		})
	}

	return s, nil
}

// TODO:
// need to dynamically load a configuration file.
// For now, we will limit the implementation to just loading the file at startup.
// Run starts the proxy server.
func (g *Gondola) Run() {
	// TODO: do health check for upstreams.

	g.logger.Info("Runing server on port " + g.config.Proxy.Port + "...")

	go func() {
		if err := g.server.ListenAndServe(); err != http.ErrServerClosed {
			g.logger.Error("Server stopped with error: " + err.Error())
			return
		}
	}()

	q := make(chan os.Signal, 1)
	signal.Notify(q, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	<-q

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(g.config.Proxy.ShutdownTimeout)*time.Millisecond)
	defer cancel()
	if err := g.server.Shutdown(ctx); err != nil {
		g.logger.Error("Server stopped with error: " + err.Error())
		return
	}

	g.logger.Info("Server stopped gracefully")
}
