package gondola

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFormatAccessLog(t *testing.T) {
	attrs := map[string]any{
		"remote_addr":     "192.0.2.1",
		"method":          "GET",
		"request_uri":     "/api/x",
		"protocol":        "HTTP/1.1",
		"status_code":     200,
		"body_bytes_sent": int64(123),
		"referer":         "https://example.com",
		"user_agent":      "curl/8",
		"request_time":    0.123,
	}

	common := formatAccessLog(logFormatCommon, "", attrs, "tid")
	if !strings.Contains(common, `"GET /api/x HTTP/1.1" 200 123`) {
		t.Errorf("common: %q", common)
	}

	combined := formatAccessLog(logFormatCombined, "", attrs, "tid")
	if !strings.Contains(combined, `"https://example.com" "curl/8"`) {
		t.Errorf("combined: %q", combined)
	}

	custom := formatAccessLog(logFormatCustom, "${method} ${uri} ${status} ${trace_id}", attrs, "tid")
	if custom != "GET /api/x 200 tid" {
		t.Errorf("custom: %q", custom)
	}

	if got := formatAccessLog(logFormatJSON, "", attrs, "tid"); got != "" {
		t.Errorf("json format should produce no text line, got %q", got)
	}
}

func TestLoggerAccessLogFormat(t *testing.T) {
	path := filepath.Join(t.TempDir(), "access.log")
	logger, err := NewLogger(slog.LevelInfo, path, logFormatCommon, "")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	ctx := WithTraceID(context.Background())
	logger.InfoContext(ctx, accessLogMsg,
		slog.String("remote_addr", "192.0.2.9"),
		slog.String("method", "GET"),
		slog.String("request_uri", "/"),
		slog.String("protocol", "HTTP/1.1"),
		slog.Int("status_code", 200),
		slog.Int64("body_bytes_sent", 5),
	)
	// Non-access records must remain JSON regardless of the access log format.
	logger.InfoContext(ctx, "hello", slog.String("k", "v"))
	if err := logger.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	if !strings.Contains(s, `"GET / HTTP/1.1" 200 5`) {
		t.Errorf("expected common access log line, got %q", s)
	}
	if !strings.Contains(s, `"msg":"hello"`) {
		t.Errorf("expected JSON application log, got %q", s)
	}
}
