package gondola

import (
	"reflect"
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	data := `
upstreams:
  - name: backend1
    address: http://backend1:8081
  - name: backend2
    address: http://backend2:8082
`

	expected := []Upstream{
		{
			Name:    "backend1",
			Address: "http://backend1:8081",
		},
		{
			Name:    "backend2",
			Address: "http://backend2:8082",
		},
	}

	actual := &Config{}
	if _, err := actual.Load(strings.NewReader(data)); err != nil {
		t.Error(err)
	}

	for i, b := range actual.Upstreams {
		if !reflect.DeepEqual(expected[i], b) {
			t.Errorf("Expected %+v, got %+v", expected[i], b)
		}
	}
}
