package gondola

import (
	"errors"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// waitForListen blocks until addr accepts a TCP connection or timeout elapses.
func waitForListen(addr string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, 100*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return true
		}
		time.Sleep(20 * time.Millisecond)
	}
	return false
}

func TestGondolaShutdown(t *testing.T) {
	data := `
proxy:
  port: "18099"
  shutdown_timeout: 2000
upstreams:
  - host_name: backend.local
    target: http://localhost:65535
log_level: error
`
	g, err := NewGondola(strings.NewReader(data))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- g.server.ListenAndServe()
	}()

	if !waitForListen("127.0.0.1:18099", 2*time.Second) {
		t.Fatal("server did not start listening in time")
	}

	if err := g.Shutdown(); err != nil {
		t.Fatalf("Expected no error on shutdown, got %v", err)
	}

	if err := <-errCh; err != nil && !errors.Is(err, http.ErrServerClosed) {
		t.Errorf("Expected ErrServerClosed, got %v", err)
	}
}

func TestGondolaReloadNoConfigPath(t *testing.T) {
	g, err := NewGondola(strings.NewReader("proxy:\n  port: \"8080\"\n"))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if err := g.reload(); err != nil {
		t.Errorf("Expected nil reload without config path, got %v", err)
	}
}

func TestGondolaReload(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	cfg := `
proxy:
  port: "18098"
upstreams:
  - host_name: a.local
    target: http://localhost:9001
log_level: error
`
	if err := os.WriteFile(path, []byte(cfg), 0o600); err != nil {
		t.Fatal(err)
	}

	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	g, err := NewGondola(f, WithConfigPath(path))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	cfg2 := `
proxy:
  port: "18098"
upstreams:
  - host_name: b.local
    target: http://localhost:9002
log_level: error
`
	if err := os.WriteFile(path, []byte(cfg2), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := g.reload(); err != nil {
		t.Fatalf("Expected no error on reload, got %v", err)
	}
	if g.config.Upstreams[0].HostName != "b.local" {
		t.Errorf("Expected reloaded host b.local, got %q", g.config.Upstreams[0].HostName)
	}

	// Reloading an invalid configuration should return an error.
	if err := os.WriteFile(path, []byte("proxy:\n  port: \"1\"\nupstreams:\n  - host_name: x\n    target: \"://\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := g.reload(); err == nil {
		t.Error("Expected error reloading invalid config")
	}
}

func TestWithLogLevelOption(t *testing.T) {
	g, err := NewGondola(strings.NewReader("proxy:\n  port: \"8080\"\nlog_level: error\n"), WithLogLevel("debug"))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if g.logLevel != "debug" {
		t.Errorf("Expected log level override debug, got %q", g.logLevel)
	}
}
