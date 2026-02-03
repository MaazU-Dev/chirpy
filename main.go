package main

import (
	"net/http"
)

func main() {
	rootFileDir := "."
	port := "8080"
	mux := http.NewServeMux()
	mux.Handle("/app/", http.StripPrefix("/app/", http.FileServer(http.Dir(rootFileDir))))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(200)
		msg := "OK"
		w.Write([]byte(msg))

	})
	s := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	s.ListenAndServe()
}
