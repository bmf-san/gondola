package main

import (
	"log"
	"net/http"
	"time"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Print("backend2 sleep 2 seconds")
		time.Sleep(5 * time.Second)
		log.Print(r.Header)
		w.Write([]byte("Hello from backend2"))
	})
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("backend2 is OK"))
	})
	if err := http.ListenAndServe(":8082", mux); err != nil {
		log.Fatal(err)
	}
}
