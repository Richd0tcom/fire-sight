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

type FunctionStats struct {
	Name         string
	LineStart    int
	LineEnd      int
	TotalChanges int
	LastModified time.Time
	ChangesByDay map[int]int // days ago -> change count
}

type FileAnalysis struct {
	Path      string
	Language  string
	Functions []*FunctionStats
}


type AnalysisResult struct {
	RepoID      string    `json:"repoId"`
	RepoURL     string    `json:"repoUrl"`
	Branch      string    `json:"branch"`
	AnalyzedAt  time.Time `json:"analyzedAt"`
	CommitCount int       `json:"commitCount"`

	//this is likely a map of path -> stats
	FileStats map[string]*FileChangeStats `json:"fileStats"`
	FileFunctionAnalyses  map[string]*FileAnalysis `json:"fileFunctionAnalyses"`

	Status        string `json:"status"`
	TimeRangeDays int    `json:"timeRangeDays"`
}

type HeatScore struct {
	Path          string        `json:"path"`
	Score         float64       `json:"score"` // 0-100
	ChangeFreq    float64       `json:"changeFreq"` //TODO: compare with int
	DaysSinceEdit int           `json:"daysSinceEdit"`
	TotalFileChanges int `json:"totalFileChanges"`
}


// FileNode represents a file or folder in the tree structure
type FileNode struct {
	ID              string          `json:"id"`
	Name            string          `json:"name"`
	Path            string          `json:"path"`
	Type            FileNodeType    `json:"type"` // "file" | "folder"
	Extension       string          `json:"extension,omitempty"`
	Size            int64           `json:"size"`
	LinesOfCode     int             `json:"linesOfCode"`
	LastModified    time.Time       `json:"lastModified"`
	HeatScore       *HeatScore      `json:"heatScore"`
	Functions       []*FunctionNode `json:"functions,omitempty"`
	Children        []*FileNode     `json:"children,omitempty"`
	FileCount       int             `json:"fileCount,omitempty"` // For folders: total files inside
}

// FunctionNode represents a function/method within a file (Milestone 2)
type FunctionNode struct {
	Name            string    `json:"name"`
	LineStart       int       `json:"lineStart"`
	LineEnd         int       `json:"lineEnd"`
	LastModified    time.Time `json:"lastModified"`
	HeatScore       *HeatScore   `json:"heatScore"`
	IsDeadCode      bool      `json:"isDeadCode"`
}

type Changes struct {
	FilePath string
}

// API Request/Response types
type AnalyzeRequest struct {
	RepoURL       string `json:"repoUrl"`
	Branch        string `json:"branch"`               // default: "main"
	TimeRangeDays int    `json:"timeRangeDays"`      // default: 180
	AuthToken     string `json:"authToken,omitempty"` // for private repos
}

type AnalyzeResponse struct {
	RepoID    string      `json:"repoId"`
	Status    string      `json:"status"` // "complete" | "error"
	FileStats []HeatScore `json:"fileStats"`
	FileTree  *FileNode   `json:"fileTree,omitempty"` // NEW: Hierarchical tree
	Analyzed  int         `json:"analyzedFiles"`
	Error     string      `json:"error,omitempty"`
	Duration  string      `json:"duration"`
}
