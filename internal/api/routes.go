package api

import (
	"github.com/gorilla/mux"
)

func SetupRoutes(h * Handler) *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/analyze", h.AnalyzeRepo).Methods("POST")
	r.HandleFunc("/health", h.HealthCheck).Methods("GET")

	return r
}