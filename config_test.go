package gondola

import (
	"reflect"
	"strings"
	"testing"
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
upstreams:
  - host_name: backend1.local
    target: http://backend1:8081
  - host_name: backend2.local
    target: http://backend2:8082
log_level: -4
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
					Path: "/public/",
					Dir:  "testdata/public",
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
		4,
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
