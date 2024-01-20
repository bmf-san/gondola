package main

import (
	"log"
	"net/http"
	"time"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Print("backend1 sleep 1 seconds")
		time.Sleep(3 * time.Second)
		log.Print(r.Header)
		w.Write([]byte("Hello from backend1"))
	})
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("backend1 is OK"))
	})
	if err := http.ListenAndServe(":8081", mux); err != nil {
		log.Fatal(err)
	}
}
