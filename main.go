package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func main() {
	rootFileDir := "."
	port := "8080"
	mux := http.NewServeMux()
	var config apiConfig
	mux.Handle("/app/", http.StripPrefix("/app/", config.middlewareMetricsInc(http.FileServer(http.Dir(rootFileDir)))))
	mux.HandleFunc("/metrics", config.handlerMetrics)
	mux.HandleFunc("/reset", config.handlerReset)
	mux.HandleFunc("/health", handlerReadiness)
	s := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	s.ListenAndServe()
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Hits: %d", cfg.fileserverHits.Load())))
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	},
	)
}
