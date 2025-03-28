package main

import (
	"net/http"
	"strconv"
	"sync/atomic"
)

func main() {
	const filepathRoot = "."
	const port = "8080"
	config := &apiConfig{}

	mux := http.NewServeMux()
	mux.Handle("/app/", http.StripPrefix("/app", config.middlewareMetricsInc(http.FileServer(http.Dir(filepathRoot)))))

	mux.HandleFunc("GET /healthz", handlerReadiness)
	mux.HandleFunc("GET /metrics", config.handlerMetrics)
	mux.HandleFunc("POST /reset", config.handlerMetricsReset)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	server.ListenAndServe()
}

type apiConfig struct {
	fileserverHits atomic.Int32
}

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		cfg.fileserverHits.Add(1)
	},
	)
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	numHits := strconv.Itoa(int(cfg.fileserverHits.Load()))
	w.Write([]byte("Hits: " + numHits))
}

func (cfg *apiConfig) handlerMetricsReset(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits reset to 0"))
}
