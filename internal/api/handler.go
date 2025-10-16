package api

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/richd0tcom/fire-sight/internal/analyzer"
	"github.com/richd0tcom/fire-sight/internal/models"
)



type Handler struct {
	gitAnalyzer    *analyzer.GitAnalyzer
	heatCalculator *analyzer.HeatCalculator
	timeout        time.Duration
}

func NewHandler(gitAnalyzer *analyzer.GitAnalyzer, heatCalculator *analyzer.HeatCalculator) *Handler {
	return &Handler{
		gitAnalyzer:    gitAnalyzer,
		heatCalculator: heatCalculator,
		timeout:        5 * time.Minute, // Max time for repo analysis
	}
}

// AnalyzeRepo handles POST /api/analyze
func (h *Handler) AnalyzeRepo(w http.ResponseWriter, r *http.Request) {
	// Parse request
	var req models.AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate
	if req.RepoURL == "" {
		h.respondError(w, http.StatusBadRequest, "repo_url is required")
		return
	}

	// Set defaults
	if req.Branch == "" {
		req.Branch = "main"
	}
	if req.TimeRangeDays == 0 {
		req.TimeRangeDays = 180
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), h.timeout)
	defer cancel()

	// Start timer for performance tracking
	startTime := time.Now()

	// STEP 1: Analyze git history
	opts := models.AnalyzeOptions{
		Branch:        req.Branch,
		TimeRangeDays: req.TimeRangeDays,
		AuthToken:     req.AuthToken,
	}

	result, err := h.gitAnalyzer.AnalyzeRepository(ctx, req.RepoURL, opts)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, fmt.Sprintf("Analysis failed: %v", err))
		return
	}

	// STEP 2: Calculate heat scores
	heatScores := h.heatCalculator.CalculateHeatScores(result)

	// Generate repo ID (for future caching)
	repoID := h.generateRepoID(req.RepoURL, req.Branch)

	// Build response
	response := models.AnalyzeResponse{
		RepoID:    repoID,
		Status:    "complete",
		FileStats: heatScores,
		Analyzed:  len(heatScores),
		Duration:  time.Since(startTime).String(),
	}

	h.respondJSON(w, http.StatusOK, response)
}

// HealthCheck handles GET /health
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	}
	h.respondJSON(w, http.StatusOK, response)
}

// Helper: Generate deterministic repo ID
func (h *Handler) generateRepoID(repoURL, branch string) string {
	data := fmt.Sprintf("%s:%s", repoURL, branch)
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash[:8]) // First 8 bytes = 16 hex chars
}

// Helper: Send JSON response
func (h *Handler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// Helper: Send error response
func (h *Handler) respondError(w http.ResponseWriter, status int, message string) {
	response := models.AnalyzeResponse{
		Status: "error",
		Error:  message,
	}
	h.respondJSON(w, status, response)
}