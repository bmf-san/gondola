package gondola

import (
	"bytes"
	"errors"
	"log/slog"
	"reflect"
	"strings"
	"testing"
	"testing/iotest"
	"time"
)

func TestIsEnableTLS(t *testing.T) {
	cases := []struct {
		name     string
		item     *Proxy
		expected bool
	}{
		{
			name:     "TLSCertPath and TLSKeyPath are empty",
			item:     &Proxy{},
			expected: false,
		},
		{
			name:     "TLSCertPath is empty",
			item:     &Proxy{TLSKeyPath: "key"},
			expected: false,
		},
		{
			name:     "TLSKeyPath is empty",
			item:     &Proxy{TLSCertPath: "cert"},
			expected: false,
		},
		{
			name:     "TLSCertPath and TLSKeyPath are not empty",
			item:     &Proxy{TLSCertPath: "cert", TLSKeyPath: "key"},
			expected: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual := c.item.IsEnableTLS()
			if actual != c.expected {
				t.Fatalf("Expected %v, got %v", c.expected, actual)
			}
		})
	}
}

func TestLoad(t *testing.T) {
	t.Setenv("PORT", "8080")
	data := `
proxy:
  port: ${PORT}
  read_header_timeout: 2000
  shutdown_timeout: 3000
  tls_cert_path: /path/to/cert
  tls_key_path: /path/to/key
  static_files:
    - path: /public/
      dir: testdata/public
      fallback_path: custom_404.html
upstreams:
  - host_name: backend1.local
    target: http://backend1:8081
  - host_name: backend2.local
    target: http://backend2:8082
log_level: debug
`

	expected := &Config{
		Proxy{
			Port:              "8080",
			ReadHeaderTimeout: 2000,
			ShutdownTimeout:   3000,
			TLSCertPath:       "/path/to/cert",
			TLSKeyPath:        "/path/to/key",
			StaticFiles: []StaticFile{
				{
					Path:         "/public/",
					Dir:          "testdata/public",
					FallbackPath: "custom_404.html",
				},
			},
		},
		[]Upstream{
			{
				HostName: "backend1.local",
				Target:   "http://backend1:8081",
			},
			{
				HostName: "backend2.local",
				Target:   "http://backend2:8082",
			},
		},
		"debug",
		"",
		"",
	}

	actual := &Config{}
	if _, err := actual.Load(strings.NewReader(data)); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !reflect.DeepEqual(expected.Proxy, actual.Proxy) {
		t.Fatalf("Expected %+v, got %+v", expected.Proxy, actual.Proxy)
	}

	for i, b := range actual.Upstreams {
		if !reflect.DeepEqual(expected.Upstreams[i], b) {
			t.Fatalf("Expected %+v, got %+v", expected.Upstreams[i], b)
		}
	}
}

func TestLoadReadAllError(t *testing.T) {
	reader := iotest.ErrReader(errors.New("error"))
	var c Config
	_, err := c.Load(reader)
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}
}

func TestLoadUnmarshalError(t *testing.T) {
	data := ":\n"
	reader := bytes.NewBufferString(data)
	var c Config
	_, err := c.Load(reader)
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}
}

func TestParseLevel(t *testing.T) {
	cases := map[string]slog.Level{
		"debug":   slog.LevelDebug,
		"DEBUG":   slog.LevelDebug,
		"info":    slog.LevelInfo,
		"warn":    slog.LevelWarn,
		"warning": slog.LevelWarn,
		"error":   slog.LevelError,
		"":        slog.LevelInfo,
		"unknown": slog.LevelInfo,
	}
	for in, want := range cases {
		if got := ParseLevel(in); got != want {
			t.Errorf("ParseLevel(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestMsToDuration(t *testing.T) {
	if got := msToDuration(0); got != 0 {
		t.Errorf("msToDuration(0) = %v, want 0", got)
	}
	if got := msToDuration(-5); got != 0 {
		t.Errorf("msToDuration(-5) = %v, want 0", got)
	}
	if got := msToDuration(1500); got != 1500*time.Millisecond {
		t.Errorf("msToDuration(1500) = %v, want 1.5s", got)
	}
}
