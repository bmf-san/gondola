package gondola

import (
	"reflect"
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	data := `
proxy:
  port: 80
  read_header_timeout: 20
  shutdown_timeout: 3000
upstreams:
  - host_name: backend1.local
    target: http://backend1:8081
  - host_name: backend2.local
    target: http://backend2:8082
`

	expected := &Config{
		Proxy{
			Port:              "80",
			ReadHeaderTimeout: 20,
			ShutdownTimeout:   3000,
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
	}

	actual := &Config{}
	if _, err := actual.Load(strings.NewReader(data)); err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(expected.Proxy, actual.Proxy) {
		t.Errorf("Expected %+v, got %+v", expected.Proxy, actual.Proxy)
	}

	for i, b := range actual.Upstreams {
		if !reflect.DeepEqual(expected.Upstreams[i], b) {
			t.Errorf("Expected %+v, got %+v", expected.Upstreams[i], b)
		}
	}
}
