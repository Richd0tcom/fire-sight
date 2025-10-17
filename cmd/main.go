package main

import (
	"log"
	"net/http"
	"os"

	"github.com/richd0tcom/fire-sight/internal/analyzer"
	"github.com/richd0tcom/fire-sight/internal/api"
)



func main() {

	port := getEnv("PORT", "8080")
	tempDir := getEnv("TEMP_DIR", "./tmp/dead-code-heatmap")

	if err := os.MkdirAll(tempDir, 0755); err != nil {
		log.Fatalf("Failed to create temp directory: %v", err)
	}

	
	gitAnalyzer := analyzer.NewGitAnalyzer(tempDir)
	heatCalculator := analyzer.NewHeatCalculator()
	handler := api.NewHandler(gitAnalyzer, heatCalculator)

	router := api.SetupRoutes(handler)

	log.Printf("Dead Code Heatmap API starting on port %s", port)
	log.Printf("Using temp directory: %s", tempDir)
	log.Printf("Ready to analyze repositories!")
	
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}