package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/richd0tcom/fire-sight/pkg"
)

func SetupRoutes(h *Handler) *mux.Router {
	r := mux.NewRouter()

	r.Use(corsMiddleware)

	r.HandleFunc("/analyze", h.AnalyzeRepo).Methods("POST")
	r.HandleFunc("/health", h.HealthCheck).Methods("GET")

	r.Methods(http.MethodOptions).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	return r
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := pkg.GetEnv("FRONTEND_URL", "http://localhost:8080")

		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
