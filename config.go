package gondola

import (
	"io"

	"gopkg.in/yaml.v3"
)

// Upstream is a struct that represents a backend server.
type Upstream struct {
	Name    string `yaml:"name"`
	Address string `yaml:"address"`
}

// Config is a struct that represents the configuration of the proxy.
type Config struct {
	Upstreams []Upstream `yaml:"upstreams"`
}

// Load reads the configuration from a reader and returns a Config struct.
func (c *Config) Load(reader io.Reader) (*Config, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return c, nil
}
