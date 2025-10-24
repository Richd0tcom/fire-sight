package models

import "time"

type FileNodeType string

const (
	FileNodeTypeFile   FileNodeType = "file"
	FileNodeTypeFolder FileNodeType = "folder"
)

type AnalyzeOptions struct {
	Branch        string
	TimeRangeDays int

	AuthToken string
}

type FileChangeStats struct {
	FilePath     string
	TotalChanges int
	LastModified time.Time
	ChangesByDay map[int]int

	//map of authors and commit count
	UniqueAuthors map[string]int
	FirstSeen     time.Time
}



type AnalysisResult struct {
	RepoID      string    `json:"repo_id"`
	RepoURL     string    `json:"repo_url"`
	Branch      string    `json:"branch"`
	AnalyzedAt  time.Time `json:"analyzed_at"`
	CommitCount int       `json:"commit_count"`

	//this is likely a map of path -> stats
	FileStats map[string]*FileChangeStats `json:"file_stats"`

	Status        string `json:"status"`
	TimeRangeDays int    `json:"time_range_days"`
}

type HeatScore struct {
	Path          string        `json:"path"`
	Score         float64       `json:"score"` // 0-100
	ChangeFreq    float64       `json:"change_freq"` //TODO: compare with int
	DaysSinceEdit int           `json:"days_since_edit"`
	TotalFileChanges int `json:"total_file_changes"`
}


// FileNode represents a file or folder in the tree structure
type FileNode struct {
	ID              string          `json:"id"`
	Name            string          `json:"name"`
	Path            string          `json:"path"`
	Type            FileNodeType    `json:"type"` // "file" | "folder"
	Extension       string          `json:"extension,omitempty"`
	Size            int64           `json:"size"`
	LinesOfCode     int             `json:"lines_of_code"`
	LastModified    time.Time       `json:"last_modified"`
	HeatScore       *HeatScore      `json:"heat_score"`
	Functions       []*FunctionNode `json:"functions,omitempty"`
	Children        []*FileNode     `json:"children,omitempty"`
	FileCount       int             `json:"file_count,omitempty"` // For folders: total files inside
}

// FunctionNode represents a function/method within a file (Milestone 2)
type FunctionNode struct {
	Name            string    `json:"name"`
	LineStart       int       `json:"line_start"`
	LineEnd         int       `json:"line_end"`
	LastModified    time.Time `json:"last_modified"`
	HeatScore       *HeatScore   `json:"heat_score"`
	IsDeadCode      bool      `json:"is_dead_code"`
}

type Changes struct {
	FilePath string
}

// API Request/Response types
type AnalyzeRequest struct {
	RepoURL       string `json:"repo_url"`
	Branch        string `json:"branch"`               // default: "main"
	TimeRangeDays int    `json:"time_range_days"`      // default: 180
	AuthToken     string `json:"auth_token,omitempty"` // for private repos
}

type AnalyzeResponse struct {
	RepoID    string      `json:"repo_id"`
	Status    string      `json:"status"` // "complete" | "error"
	FileStats []HeatScore `json:"file_stats"`
	FileTree  *FileNode   `json:"file_tree,omitempty"` // NEW: Hierarchical tree
	Analyzed  int         `json:"analyzed_files"`
	Error     string      `json:"error,omitempty"`
	Duration  string      `json:"duration"`
}
