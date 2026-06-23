package gondola

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/google/uuid"
)

// Logger wraps slog.Logger together with the underlying writer so that access
// logs can be reopened after rotation (e.g. on SIGUSR1).
type Logger struct {
	*slog.Logger
	writer *reopenableWriter
}

// NewLogger creates a logger at the given level. When path is empty logs are
// written to stdout; otherwise they are appended to the file at path, which can
// be reopened via Reopen. Application logs are always JSON; access logs honor
// format ("json", "common", "combined", "custom"), with customFormat used when
// format is "custom".
func NewLogger(level slog.Level, path, format, customFormat string) (*Logger, error) {
	w, err := newReopenableWriter(path)
	if err != nil {
		return nil, err
	}
	inner := TraceIDHandler{slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level: level,
	})}
	handler := &accessLogHandler{
		inner:        inner,
		w:            w,
		format:       format,
		customFormat: customFormat,
	}
	return &Logger{
		Logger: slog.New(handler),
		writer: w,
	}, nil
}

// Reopen reopens the underlying log file. It is a no-op when logging to stdout.
func (l *Logger) Reopen() error {
	return l.writer.reopen()
}

// Close closes the underlying log file. It is a no-op when logging to stdout.
func (l *Logger) Close() error {
	return l.writer.Close()
}

// accessLogHandler is a slog.Handler that renders access log records
// ("access_log") in the configured text format (common/combined/custom),
// delegating everything else (and the "json" format) to the wrapped handler.
type accessLogHandler struct {
	inner        slog.Handler
	w            io.Writer
	format       string
	customFormat string
}

func (h *accessLogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

func (h *accessLogHandler) Handle(ctx context.Context, r slog.Record) error {
	if h.format == "" || h.format == logFormatJSON || r.Message != accessLogMsg {
		return h.inner.Handle(ctx, r)
	}
	attrs := make(map[string]any, r.NumAttrs())
	r.Attrs(func(a slog.Attr) bool {
		attrs[a.Key] = a.Value.Any()
		return true
	})
	line := formatAccessLog(h.format, h.customFormat, attrs, GetTraceID(ctx))
	_, err := io.WriteString(h.w, line+"\n")
	return err
}

func (h *accessLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &accessLogHandler{inner: h.inner.WithAttrs(attrs), w: h.w, format: h.format, customFormat: h.customFormat}
}

func (h *accessLogHandler) WithGroup(name string) slog.Handler {
	return &accessLogHandler{inner: h.inner.WithGroup(name), w: h.w, format: h.format, customFormat: h.customFormat}
}

// reopenableWriter is an io.Writer backed by either stdout or a file that can
// be reopened (closing the old descriptor) to support external log rotation.
type reopenableWriter struct {
	path string
	mu   sync.Mutex
	f    *os.File
	w    io.Writer
}

func newReopenableWriter(path string) (*reopenableWriter, error) {
	rw := &reopenableWriter{path: path}
	if path == "" {
		rw.w = os.Stdout
		return rw, nil
	}
	if err := rw.reopen(); err != nil {
		return nil, err
	}
	return rw, nil
}

func (rw *reopenableWriter) Write(p []byte) (int, error) {
	rw.mu.Lock()
	defer rw.mu.Unlock()
	return rw.w.Write(p)
}

func (rw *reopenableWriter) reopen() error {
	if rw.path == "" {
		return nil // stdout: nothing to reopen
	}
	// #nosec G304 -- path is operator-provided configuration, not user input.
	f, err := os.OpenFile(filepath.Clean(rw.path), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}
	rw.mu.Lock()
	old := rw.f
	rw.f = f
	rw.w = f
	rw.mu.Unlock()
	if old != nil {
		_ = old.Close()
	}
	return nil
}

func (rw *reopenableWriter) Close() error {
	rw.mu.Lock()
	defer rw.mu.Unlock()
	if rw.f != nil {
		err := rw.f.Close()
		rw.f = nil
		return err
	}
	return nil // stdout: never close
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
	tid := GetTraceID(ctx)
	if tid != "" {
		r.AddAttrs(slog.String("trace_id", tid))
	}
	return t.Handler.Handle(ctx, r)
}
