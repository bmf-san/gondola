package gondola

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

func New() {
	backend, _ := url.Parse("http://backend:8081")
	proxy := httputil.NewSingleHostReverseProxy(backend)
	http.Handle("/", proxy)
	http.ListenAndServe(":8080", nil)
}
