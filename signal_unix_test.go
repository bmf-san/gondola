//go:build unix

package gondola

import (
	"strings"
	"syscall"
	"testing"
)

func TestHandleSignal(t *testing.T) {
	g, err := NewGondola(strings.NewReader("proxy:\n  port: \"8080\"\n"))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !g.handleSignal(syscall.SIGTERM) {
		t.Error("SIGTERM should request shutdown")
	}
	if !g.handleSignal(syscall.SIGINT) {
		t.Error("SIGINT should request shutdown")
	}
	// SIGHUP reloads (no config path here -> warns) and keeps running.
	if g.handleSignal(syscall.SIGHUP) {
		t.Error("SIGHUP should not request shutdown")
	}
	// SIGUSR1 reopens logs (stdout -> no-op) and keeps running.
	if g.handleSignal(syscall.SIGUSR1) {
		t.Error("SIGUSR1 should not request shutdown")
	}
}
