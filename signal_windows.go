//go:build windows

package gondola

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

// notifySignals registers the OS signals gondola handles on Windows.
// SIGHUP and SIGUSR1 are not available on Windows, so configuration reload and
// log reopen are unsupported there; only interrupt/terminate are handled.
func notifySignals(ch chan<- os.Signal) {
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
}

// handleSignal processes a received signal, returning true when the server
// should shut down.
func (g *Gondola) handleSignal(sig os.Signal) bool {
	g.logger.Info("received shutdown signal", slog.String("signal", sig.String()))
	return true
}
