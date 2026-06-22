package gondola

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Default values applied when the configuration omits a field.
// All timeout values are expressed in milliseconds to match the config file.
const (
	defaultPort                = "8080"
	defaultReadHeaderTimeoutMS = 2000
	defaultReadTimeoutMS       = 30000
	defaultWriteTimeoutMS      = 30000
	defaultIdleTimeoutMS       = 60000
	defaultShutdownTimeoutMS   = 30000
	defaultUpstreamTimeoutMS   = 30000
	defaultMaxHeaderBytes      = 1 << 20 // 1 MiB
	defaultLogLevel            = "info"
)

// Proxy is a struct that represents the proxy server.
// All *Timeout fields are expressed in milliseconds.
type Proxy struct {
	Port              string       `yaml:"port"`
	ReadHeaderTimeout int          `yaml:"read_header_timeout"`
	ReadTimeout       int          `yaml:"read_timeout"`
	WriteTimeout      int          `yaml:"write_timeout"`
	IdleTimeout       int          `yaml:"idle_timeout"`
	ShutdownTimeout   int          `yaml:"shutdown_timeout"`
	MaxHeaderBytes    int          `yaml:"max_header_bytes"`
	TLSCertPath       string       `yaml:"tls_cert_path"`
	TLSKeyPath        string       `yaml:"tls_key_path"`
	LogFile           string       `yaml:"log_file"`
	StaticFiles       []StaticFile `yaml:"static_files"`
}

// StaticFile is a struct that represents a static file configuration.
// FallbackPath (relative to Dir) is served when the requested file is missing
// or the request targets a directory without an index.html.
type StaticFile struct {
	Path         string `yaml:"path"`
	Dir          string `yaml:"dir"`
	FallbackPath string `yaml:"fallback_path"` // Path to fallback file when requested file is not found
}

// IsEnableTLS returns true if the proxy server is configured to use TLS.
func (p *Proxy) IsEnableTLS() bool {
	return p.TLSCertPath != "" && p.TLSKeyPath != ""
}

// Upstream is a struct that represents a backend server.
// HostName is matched against the request Host header.
// Target is the URL that matching requests are forwarded to.
// ReadTimeout/WriteTimeout (milliseconds) bound the time spent waiting for
// upstream response headers and establishing the connection, respectively.
type Upstream struct {
	HostName     string `yaml:"host_name"`
	Target       string `yaml:"target"`
	ReadTimeout  int    `yaml:"read_timeout"`
	WriteTimeout int    `yaml:"write_timeout"`
}

// Config is a struct that represents the configuration of the proxy.
type Config struct {
	Proxy     Proxy      `yaml:"proxy"`
	Upstreams []Upstream `yaml:"upstreams"`
	LogLevel  string     `yaml:"log_level"` // debug, info, warn, error
}

// Load reads the configuration from a reader, expands environment variables of
// the form $VAR or ${VAR}, and parses it as YAML.
func (c *Config) Load(reader io.Reader) (*Config, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	data = []byte(os.ExpandEnv(string(data)))
	if err := yaml.Unmarshal(data, c); err != nil {
		return nil, err
	}
	return c, nil
}

// setDefaults fills zero-valued fields with sensible defaults. It is
// idempotent and safe to call multiple times.
func (c *Config) setDefaults() {
	if c.Proxy.Port == "" {
		c.Proxy.Port = defaultPort
	}
	if c.Proxy.ReadHeaderTimeout == 0 {
		c.Proxy.ReadHeaderTimeout = defaultReadHeaderTimeoutMS
	}
	if c.Proxy.ReadTimeout == 0 {
		c.Proxy.ReadTimeout = defaultReadTimeoutMS
	}
	if c.Proxy.WriteTimeout == 0 {
		c.Proxy.WriteTimeout = defaultWriteTimeoutMS
	}
	if c.Proxy.IdleTimeout == 0 {
		c.Proxy.IdleTimeout = defaultIdleTimeoutMS
	}
	if c.Proxy.ShutdownTimeout == 0 {
		c.Proxy.ShutdownTimeout = defaultShutdownTimeoutMS
	}
	if c.Proxy.MaxHeaderBytes == 0 {
		c.Proxy.MaxHeaderBytes = defaultMaxHeaderBytes
	}
	if c.LogLevel == "" {
		c.LogLevel = defaultLogLevel
	}
	for i := range c.Upstreams {
		if c.Upstreams[i].ReadTimeout == 0 {
			c.Upstreams[i].ReadTimeout = defaultUpstreamTimeoutMS
		}
		if c.Upstreams[i].WriteTimeout == 0 {
			c.Upstreams[i].WriteTimeout = defaultUpstreamTimeoutMS
		}
	}
}

// SlogLevel converts the configured log level string to a slog.Level.
func (c *Config) SlogLevel() slog.Level {
	return ParseLevel(c.LogLevel)
}

// ParseLevel converts a log level string (debug, info, warn, error) to a
// slog.Level. Unknown values fall back to info.
func ParseLevel(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// msToDuration converts a millisecond count to a time.Duration. Non-positive
// values yield a zero duration (interpreted as "no timeout" by net/http).
func msToDuration(ms int) time.Duration {
	if ms <= 0 {
		return 0
	}
	return time.Duration(ms) * time.Millisecond
}

// validateUpstreams ensures every upstream target is a usable absolute URL.
func validateUpstreams(upstreams []Upstream) error {
	for _, up := range upstreams {
		if err := validateTarget(up.Target); err != nil {
			return fmt.Errorf("invalid upstream target URL %q: %w", up.Target, err)
		}
	}
	return nil
}
