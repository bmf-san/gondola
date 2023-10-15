package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello from backend1"))
	})
	if err := http.ListenAndServe(":8081", mux); err != nil {
		log.Fatal(err)
	}
}
