package main

import (
	"log"
	"net/http"
	"os"

	"github.com/richd0tcom/fire-sight/internal/analyzer"
	"github.com/richd0tcom/fire-sight/internal/api"
	"github.com/richd0tcom/fire-sight/pkg"
)



func main() {

	port := pkg.GetEnv("PORT", "8090")
	tempDir := pkg.GetEnv("TEMP_DIR", "./tmp/dead-code-heatmap")

	if err := os.MkdirAll(tempDir, 0755); err != nil {
		log.Fatalf("Failed to create temp directory: %v", err)
	}

	
	gitAnalyzer := analyzer.NewGitAnalyzer(tempDir)
	heatCalculator := analyzer.NewHeatCalculator()
	treeBuilder := analyzer.NewTreeBuilder(heatCalculator)
	handler := api.NewHandler(gitAnalyzer, treeBuilder)

	router := api.SetupRoutes(handler)

	log.Printf("Dead Code Heatmap API starting on port %s", port)
	log.Printf("Using temp directory: %s", tempDir)
	log.Printf("Ready to analyze repositories!")
	
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

