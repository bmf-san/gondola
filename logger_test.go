package gondola

import (
	"context"
	"log/slog"
	"os"
	"testing"
)

func TestNewLogger(t *testing.T) {
	level := 1
	logger := NewLogger(level)
	if logger == nil {
		t.Fatal("Expected logger to be created, but got nil")
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
