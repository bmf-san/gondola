package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello from backend2"))
	})
	if err := http.ListenAndServe(":8082", mux); err != nil {
		log.Fatal(err)
	}
}
