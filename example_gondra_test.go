package gondola_test

import (
	"strings"

	"github.com/bmf-san/gondola"
)

func ExampleNew() {
	data := `
upstreams:
  - name: backend1
    address: http://backend1:8081
  - name: backend2
    address: http://backend2:8082
`
	gondola.New(strings.NewReader(data))
	// Output:
}
