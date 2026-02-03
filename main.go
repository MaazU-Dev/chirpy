package main

import (
	"net/http"
)

func main() {
	rootFileDir := "."
	port := "8080"
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(rootFileDir)))
	s := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	s.ListenAndServe()
}
