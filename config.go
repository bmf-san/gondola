package gondola

import (
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

// Proxy is a struct that represents the proxy server.
// Port is the port that the proxy server will listen on.
// ShutdownTimeout is the timeout in milliseconds for the proxy server to shutdown.
type Proxy struct {
	Port              string       `yaml:"port"`
	ReadHeaderTimeout int          `yaml:"read_header_timeout"`
	ShutdownTimeout   int          `yaml:"shutdown_timeout"`
	TLSCertPath       string       `yaml:"tls_cert_path"`
	TLSKeyPath        string       `yaml:"tls_key_path"`
	StaticFiles       []StaticFile `yaml:"static_files"`
}

// StaticFile is a struct that represents a static file configuration.
type StaticFile struct {
	Path        string         `yaml:"path"`
	Dir         string         `yaml:"dir"`
	Fallback    string         `yaml:"fallback,omitempty"`
	DefaultFile string         `yaml:"default_file,omitempty"`
	ErrorPages  map[int]string `yaml:"error_pages,omitempty"`
}

// IsEnableTLS returns true if the proxy server is configured to use TLS.
func (p *Proxy) IsEnableTLS() bool {
	return p.TLSCertPath != "" && p.TLSKeyPath != ""
}

// Upstream is a struct that represents a backend server.
// HostName is the hostname that the proxy will listen for.
// Target is the target URL that the proxy will forward requests to.
type Upstream struct {
	HostName string `yaml:"host_name"`
	Target   string `yaml:"target"`
}

// Config is a struct that represents the configuration of the proxy.
type Config struct {
	Proxy     Proxy      `yaml:"proxy"`
	Upstreams []Upstream `yaml:"upstreams"`
	LogLevel  int        `yaml:"log_level"` // Debug:-4 Info:0 Warn:4 Error:8
}

// Load reads the configuration from a reader and returns a Config struct.
func (c *Config) Load(reader io.Reader) (*Config, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	data = []byte(os.ExpandEnv(string(data)))
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return c, nil
}
