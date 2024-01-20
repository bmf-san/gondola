package gondola

import (
	"context"
	"log/slog"
	"os"

	"github.com/google/uuid"
)

// Logger is a logger.
type Logger struct {
	*slog.Logger
}

// NewLogger creates a logger.
func NewLogger(level int) *Logger {
	handler := TraceIDHandler{slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.Level(level),
	})}
	logger := slog.New(handler)
	return &Logger{
		logger,
	}
}

// WithTraceID adds a trace ID to the context.
func WithTraceID(ctx context.Context) context.Context {
	uuid, _ := uuid.NewRandom()
	return context.WithValue(ctx, ctxTraceIDKey, uuid.String())
}

// GetTraceID returns a trace ID from the context.
func GetTraceID(ctx context.Context) string {
	tid, _ := ctx.Value(ctxTraceIDKey).(string)
	return tid
}

// TraceIDHandler is a handler for trace ID.
type TraceIDHandler struct {
	slog.Handler
}

type ctxTraceID struct{}

var ctxTraceIDKey = ctxTraceID{}

// Handle adds a trace ID to the record.
func (t TraceIDHandler) Handle(ctx context.Context, r slog.Record) error {
	tid, ok := ctx.Value(ctxTraceIDKey).(string)
	if ok {
		r.AddAttrs(slog.String("trace_id", tid))
	}
	return t.Handler.Handle(ctx, r)
}
