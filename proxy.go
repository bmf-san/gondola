package gondola

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

// LogRoundTripper is a RoundTripper that logs the request and response.
type LogRoundTripper struct {
	transport http.RoundTripper
}

// NewLogRoundTripper returns a new LogRoundTripper.
func NewLogRoundTripper(transport http.RoundTripper) *LogRoundTripper {
	return &LogRoundTripper{transport: transport}
}

// RoundTrip implements the RoundTripper interface.
// It logs the request and response.
func (lrt *LogRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	start := time.Now()

	resp, err := lrt.transport.RoundTrip(r)
	if err != nil {
		return nil, err
	}

	slog.InfoContext(r.Context(), "upstream_response", slog.Time("time", time.Now()), slog.String("client_ip", r.RemoteAddr), slog.String("req_x_forwarded_for", r.Header.Get("X-Forwarded-For")), slog.String("req_method", r.Method), slog.String("req_uri", r.RequestURI), slog.String("resp_status", resp.Status), slog.Int64("req_size", r.ContentLength), slog.Int64("resp_body_size", resp.ContentLength), slog.Float64("upstream_response_time", time.Since(start).Seconds()), slog.String("referer", r.Header.Get("referer")), slog.String("req_ua", r.UserAgent()))

	return resp, nil
}

// ProxyHandler is a http.Handler that proxies the request.
type ProxyHandler struct {
	proxy *httputil.ReverseProxy
}

// Handler implements the http.Handler interface.
// It proxies the request.
func (h *ProxyHandler) Handler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := WithTraceID(r.Context())
	defer func() {
		slog.InfoContext(ctx, "proxy_response", slog.Time("time", time.Now()), slog.String("client_ip", r.RemoteAddr), slog.String("req_x_forwarded_for", r.Header.Get("X-Forwarded-For")), slog.String("req_method", r.Method), slog.String("req_uri", r.RequestURI), slog.Int64("req_size", r.ContentLength), slog.Float64("proxy_response_time", time.Since(start).Seconds()), slog.String("referer", r.Header.Get("referer")), slog.String("req_ua", r.UserAgent()))
	}()
	h.proxy.ServeHTTP(w, r.WithContext(ctx))
}

// newServer returns a new http.Server.
func newServer(c *Config) (*http.Server, error) {
	mux := http.NewServeMux()
	for _, b := range c.Upstreams {
		pp, err := url.Parse(b.Target)
		if err != nil {
			return nil, fmt.Errorf("error parsing upstream address: %w", err)
		}

		proxy := &httputil.ReverseProxy{
			Transport: NewLogRoundTripper(http.DefaultTransport),
			Director: func(r *http.Request) {
				r.Header.Set("X-Trace-ID", GetTraceID(r.Context()))
				r.URL = pp
			},
		}
		ph := &ProxyHandler{proxy: proxy}
		mux.HandleFunc(b.HostName+"/", ph.Handler)
	}
	for _, sf := range c.Proxy.StaticFiles {
		mux.Handle(sf.Path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := WithTraceID(r.Context())
			slog.InfoContext(ctx, "static_files_response", slog.Time("time", time.Now()), slog.String("client_ip", r.RemoteAddr), slog.String("req_x_forwarded_for", r.Header.Get("X-Forwarded-For")), slog.String("req_method", r.Method), slog.String("req_uri", r.RequestURI), slog.Int64("req_size", r.ContentLength), slog.String("referer", r.Header.Get("referer")), slog.String("req_ua", r.UserAgent()))
			http.StripPrefix(sf.Path, http.FileServer(http.Dir(sf.Dir))).ServeHTTP(w, r)
		}))
	}

	s := &http.Server{
		Addr:              ":" + c.Proxy.Port,
		ReadHeaderTimeout: time.Duration(c.Proxy.ReadHeaderTimeout) * time.Millisecond,
		Handler:           mux,
	}

	return s, nil
}
