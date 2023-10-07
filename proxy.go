package gondola

import (
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func New(cfgReader io.Reader) {
	cfg := &Config{}
	if _, err := cfg.Load(cfgReader); err != nil {
		log.Fatal("Error loading config file")
	}

	for _, b := range cfg.Upstreams {
		upstream, err := url.Parse(b.Address)
		if err != nil {
			log.Fatal("Error parsing upstream address")
		}
		proxy := httputil.NewSingleHostReverseProxy(upstream)
		http.Handle("/", proxy)
		// TODO: implement virtual hosts
		break
	}

	log.Print("Starting server on port 8080")
	// TODO: implement graceful shutdown
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
