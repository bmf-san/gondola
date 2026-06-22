//go:build unix

package gondola

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

// notifySignals registers the OS signals gondola handles on Unix systems.
func notifySignals(ch chan<- os.Signal) {
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGUSR1)
}

// handleSignal processes a received signal, returning true when the server
// should shut down.
func (g *Gondola) handleSignal(sig os.Signal) bool {
	switch sig {
	case syscall.SIGHUP:
		g.logger.Info("received SIGHUP: reloading configuration")
		if err := g.reload(); err != nil {
			g.logger.Error("failed to reload configuration", slog.String("error", err.Error()))
		}
		return false
	case syscall.SIGUSR1:
		g.logger.Info("received SIGUSR1: reopening access logs")
		if err := g.logger.Reopen(); err != nil {
			g.logger.Error("failed to reopen access logs", slog.String("error", err.Error()))
		}
		return false
	default: // SIGINT, SIGTERM
		g.logger.Info("received shutdown signal", slog.String("signal", sig.String()))
		return true
	}
}
