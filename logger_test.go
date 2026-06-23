package gondola

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewLogger(t *testing.T) {
	logger, err := NewLogger(slog.LevelInfo, "", "json", "")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if logger == nil {
		t.Fatal("Expected logger to be created, but got nil")
	}
}

func TestLoggerReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "access.log")
	logger, err := NewLogger(slog.LevelInfo, path, "json", "")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	logger.Info("first")
	if err := logger.Reopen(); err != nil {
		t.Fatalf("Expected no error on reopen, got %v", err)
	}
	logger.Info("second")
	if err := logger.Close(); err != nil {
		t.Fatalf("Expected no error on close, got %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Expected to read log file, got %v", err)
	}
	if !strings.Contains(string(data), "first") || !strings.Contains(string(data), "second") {
		t.Errorf("Expected both log lines in %q", string(data))
	}
}

func TestWithAndGetTraceID(t *testing.T) {
	ctx := WithTraceID(context.Background())
	tid := GetTraceID(ctx)
	if tid == "" {
		t.Fatal("Expected trace ID to be created, but got empty string")
	}
}

func TestHandle(t *testing.T) {
	handler := TraceIDHandler{slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.Level(0),
	})}
	ctx := context.WithValue(context.Background(), ctxTraceIDKey, "12345")
	err := handler.Handle(ctx, slog.Record{Level: slog.Level(1)})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	ctx = context.Background()
	err = handler.Handle(ctx, slog.Record{Level: slog.Level(1)})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}
