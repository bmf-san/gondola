// Command standalone runs a self-contained gondola demo without Docker or
// /etc/hosts edits: it starts two in-process backends and a gondola reverse
// proxy in front of them. Run it with `go run .`.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/bmf-san/gondola"
)

const config = `
proxy:
  port: "8080"
  static_files:
    - path: /public/
      dir: ./public
      fallback_path: 404.html
upstreams:
  - host_name: backend1.local
    target: http://127.0.0.1:8081
  - host_name: backend2.local
    target: http://127.0.0.1:8082
log_level: "info"
`

const usage = `
gondola standalone demo running on http://localhost:8080  (Ctrl+C to stop)

  # virtual-host routing by Host header
  curl -H "Host: backend1.local" http://localhost:8080/
  curl -H "Host: backend2.local" http://localhost:8080/

  # static files + fallback
  curl http://localhost:8080/public/index.html
  curl http://localhost:8080/public/missing.html   # -> 404.html

`

// startBackend starts a tiny HTTP backend in a goroutine and returns its server
// so it can be shut down later.
func startBackend(addr, name string) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello from %s (path=%s)\n", name, r.URL.Path)
	})
	srv := &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("%s server error: %v", name, err)
		}
	}()
	return srv
}

func main() {
	backends := []*http.Server{
		startBackend("127.0.0.1:8081", "backend1"),
		startBackend("127.0.0.1:8082", "backend2"),
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		for _, b := range backends {
			_ = b.Shutdown(ctx)
		}
	}()

	g, err := gondola.NewGondola(strings.NewReader(config))
	if err != nil {
		log.Fatalf("failed to start gondola: %v", err)
	}

	fmt.Print(usage)

	// Run blocks until interrupted (Ctrl+C / SIGINT triggers graceful shutdown).
	if err := g.Run(); err != nil {
		log.Fatalf("gondola stopped: %v", err)
	}
}
