package gondola

import (
	"fmt"
	"net/http"
)

// Gondola is a proxy server.
type Gondola struct {
	config *Config
	server *http.Server
}

// ConfigLoadError is an error that occurs when loading the configuration.
type ConfigLoadError struct {
	Err error
}

// Error implements the error interface.
func (e *ConfigLoadError) Error() string {
	return fmt.Sprintf("error loading config: %v", e.Err)
}

// Unwrap implements the errors.Wrapper interface.
func (e *ConfigLoadError) Unwrap() error {
	return e.Err
}

// ProxyServerError is an error that occurs when creating the server.
type ProxyServerError struct {
	Err error
}

// Error implements the error interface.
func (e *ProxyServerError) Error() string {
	return fmt.Sprintf("error creating server: %v", e.Err)
}

// Unwrap implements the errors.Wrapper interface.
func (e *ProxyServerError) Unwrap() error {
	return e.Err
}
