package models

import "time"

type AnalyzeOptions struct {
	Branch string 
	TimeRangeDays int

	AuthToken string
}


type FileChangeStats struct {
	FilePath string
	TotalChanges int
	LastModified time.Time
	ChangesByDay map[int]int

	//map of authors and commit count
	UniqueAuthors map[string]int
	FirstSeen time.Time 
}


type AnalysisResult struct {
	RepoID       string
	RepoURL      string
	Branch       string
	AnalyzedAt   time.Time
	CommitCount  int

	//this is likely a map of path -> stats
	FileStats    map[string]*FileChangeStats 
	TimeRangeDays int
}

type HeatScore struct {
	Path          string
	Score         float64 // 0-100
	ChangeFreq    float64 //TODO: compare with int
	DaysSinceEdit int
}

type Changes struct {
	FilePath string
}





// API Request/Response types
type AnalyzeRequest struct {
	RepoURL       string `json:"repo_url"`
	Branch        string `json:"branch"`        // default: "main"
	TimeRangeDays int    `json:"time_range_days"` // default: 180
	AuthToken     string `json:"auth_token,omitempty"` // for private repos
}

type AnalyzeResponse struct {
	RepoID    string       `json:"repo_id"`
	Status    string       `json:"status"` // "complete" | "error"
	FileStats []HeatScore  `json:"file_stats"`
	Analyzed  int          `json:"analyzed_files"`
	Error     string       `json:"error,omitempty"`
	Duration  string       `json:"duration"`
}