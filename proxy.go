package gondola

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Load(reader io.Reader) (*Config, error) {
	cfg := &Config{}
	c, err := cfg.Load(reader)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func NewServer(cfg *Config) (*http.Server, error) {
	s := &http.Server{
		Addr:              ":" + cfg.Proxy.Port,
		ReadHeaderTimeout: time.Duration(cfg.Proxy.ReadHeaderTimeout) * time.Second,
	}

	for _, b := range cfg.Upstreams {
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
func Run(reader io.Reader) {
	cfg, err := Load(reader)
	if err != nil {
		log.Fatal(err)
	}

	// TODO: do health check for upstreams.

	srv, err := NewServer(cfg)
	if err != nil {
		log.Fatal(err)
	}

	log.Print("Runing server on port " + cfg.Proxy.Port + "...")
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	q := make(chan os.Signal, 1)
	signal.Notify(q, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	<-q

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Proxy.ShutdownTimeout)*time.Millisecond)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
	log.Print("Server stopped gracefully")
}
