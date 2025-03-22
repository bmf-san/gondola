package gondola

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"time"
)

type responseInfo struct {
	// Client info
	remoteAddr    string
	remotePort    string
	xForwardedFor string

	// Request info
	method      string
	requestURI  string
	queryString string
	host        string
	requestSize int64

	// Response info
	status         string
	bodyBytesSent  int64
	totalBytesSent int64
	responseTime   float64

	// Upstream info
	upstreamAddr   string
	upstreamStatus string
	upstreamSize   int64
	upstreamTime   float64

	// Headers
	referer   string
	userAgent string
}

type responseWriter struct {
	http.ResponseWriter
	status int
	size   int64
}

func (w *responseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *responseWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(b)
	w.size += int64(n)
	return n, err
}

// LogRoundTripper is a RoundTripper that collects information about the request and response.
type LogRoundTripper struct {
	transport http.RoundTripper
}

// NewLogRoundTripper returns a new LogRoundTripper.
func NewLogRoundTripper(transport http.RoundTripper) *LogRoundTripper {
	return &LogRoundTripper{
		transport: transport,
	}
}

// GetInfo returns the response info from the request context
func GetInfo(r *http.Request) *responseInfo {
	if info, ok := r.Context().Value(responseInfoKey{}).(*responseInfo); ok {
		return info
	}
	return nil
}

// SetInfo sets the response info in the request context
func SetInfo(r *http.Request, info *responseInfo) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), responseInfoKey{}, info))
}

type responseInfoKey struct{}

// RoundTrip implements the RoundTripper interface.
func (lrt *LogRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	start := time.Now()

	info := GetInfo(r)
	if info == nil {
		return nil, fmt.Errorf("response info not found in context")
	}

	resp, err := lrt.transport.RoundTrip(r)
	if err != nil {
		return nil, err
	}

	info.upstreamStatus = resp.Status
	info.upstreamSize = resp.ContentLength
	info.upstreamTime = time.Since(start).Seconds()
	info.upstreamAddr = r.URL.Host

	return resp, nil
}

// ProxyHandler is a http.Handler that proxies the request.
type ProxyHandler struct {
	proxy  *httputil.ReverseProxy
	logger *slog.Logger
}

// NewProxyHandler creates a new ProxyHandler.
func NewProxyHandler(proxy *httputil.ReverseProxy, logger *slog.Logger) *ProxyHandler {
	return &ProxyHandler{
		proxy:  proxy,
		logger: logger,
	}
}

// ServeHTTP implements the http.Handler interface.
func (h *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := WithTraceID(r.Context())
	rw := &responseWriter{ResponseWriter: w}

	// Create responseInfo and collect request information
	host, port := "unknown", "0"
	if r.RemoteAddr != "" {
		if h, p, err := net.SplitHostPort(r.RemoteAddr); err == nil {
			host = h
			port = p
		}
	}

	info := &responseInfo{
		remoteAddr:    host,
		remotePort:    port,
		xForwardedFor: r.Header.Get("X-Forwarded-For"),
		method:        r.Method,
		requestURI:    r.URL.String(),
		queryString:   r.URL.RawQuery,
		host:          r.Host,
		requestSize:   r.ContentLength,
		referer:       r.Header.Get("Referer"),
		userAgent:     r.Header.Get("User-Agent"),
	}

	r = r.WithContext(ctx)
	r = SetInfo(r, info)

	h.proxy.ServeHTTP(rw, r)

	info.status = http.StatusText(rw.status)
	info.bodyBytesSent = rw.size
	info.totalBytesSent = rw.size // ヘッダーサイズは現時点では計算していない
	info.responseTime = time.Since(start).Seconds()

	h.logger.InfoContext(ctx, "access_log",
		// Client info
		slog.String("remote_addr", info.remoteAddr),
		slog.String("remote_port", info.remotePort),
		slog.String("x_forwarded_for", info.xForwardedFor),

		// Request info
		slog.String("method", info.method),
		slog.String("request_uri", info.requestURI),
		slog.String("query_string", info.queryString),
		slog.String("host", info.host),
		slog.Int64("request_size", info.requestSize),

		// Response info
		slog.String("status", info.status),
		slog.Int64("body_bytes_sent", info.bodyBytesSent),
		slog.Int64("bytes_sent", info.totalBytesSent),
		slog.Float64("request_time", info.responseTime),

		// Upstream info
		slog.String("upstream_addr", info.upstreamAddr),
		slog.String("upstream_status", info.upstreamStatus),
		slog.Int64("upstream_size", info.upstreamSize),
		slog.Float64("upstream_response_time", info.upstreamTime),

		// Headers
		slog.String("referer", info.referer),
		slog.String("user_agent", info.userAgent),
	)
}
