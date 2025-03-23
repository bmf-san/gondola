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

// ErrorHandler handles HTTP errors and serves custom error pages.
type ErrorHandler struct {
	errorPages map[int]string
	fs         http.FileSystem
}

// NewErrorHandler creates a new ErrorHandler with the given error pages and file system.
func NewErrorHandler(errorPages map[int]string, fs http.FileSystem) *ErrorHandler {
	return &ErrorHandler{
		errorPages: errorPages,
		fs:         fs,
	}
}

// ServeError serves an error page for the given status code.
func (h *ErrorHandler) ServeError(w http.ResponseWriter, r *http.Request, status int) {
	if h.errorPages == nil || h.fs == nil {
		http.Error(w, http.StatusText(status), status)
		return
	}

	errorPage, exists := h.errorPages[status]
	if !exists {
		http.Error(w, http.StatusText(status), status)
		return
	}

	f, err := h.fs.Open(errorPage)
	if err != nil {
		http.Error(w, http.StatusText(status), status)
		return
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		http.Error(w, http.StatusText(status), status)
		return
	}

	w.WriteHeader(status)
	http.ServeContent(w, r, errorPage, fi.ModTime(), f)
}
