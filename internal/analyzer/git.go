package analyzer

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"

	"github.com/richd0tcom/fire-sight/internal/models"
)


type GitAnalyzer struct {
	tempDir string
}

func NewGitAnalyzer(tempDir string) *GitAnalyzer {
	return &GitAnalyzer{tempDir: tempDir}
}

func (ga *GitAnalyzer) cloneRepo(ctx context.Context, repoURL string, opts models.AnalyzeOptions) (*git.Repository, string, error) {

	repoPath := fmt.Sprintf("%s/repo-%d", ga.tempDir, time.Now().Unix())

	cloneOpts:= git.CloneOptions{
		URL: repoURL,
		Progress: nil, 
		Depth: 0, //I think we need a full clone for history
	}

	if opts.AuthToken != "" {
		cloneOpts.Auth = &http.BasicAuth{
			Username: "token", // Can be anything for token auth
			Password: opts.AuthToken,
		}
	}

	repo, err := git.PlainCloneContext(ctx, repoPath, false, &cloneOpts)
	if err != nil {
		return nil, "", err
	}

	return repo, repoPath, nil
}

func (ga *GitAnalyzer) AnalyzeRepository(ctx context.Context, repoUrl string, opts models.AnalyzeOptions) (*models.AnalysisResult, error) {

	repo, repoPath, err := ga.cloneRepo(ctx, repoUrl, opts)
	if err != nil {
		return nil, fmt.Errorf("clone failed: %w", err)
	}

	defer os.RemoveAll(repoPath) 

	fStats := make(map[string]bool)

	err = fs.WalkDir(os.DirFS(repoPath), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// Skip VCS metadata
		if path == ".git" || strings.HasPrefix(path, ".git/") {
			return fs.SkipDir
		}
	
		if d.Type().IsRegular() {
			fStats[path] = true
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk repo tree failed: %w", err)
	}

	result, err := ga.parseGitHistory(ctx, repo, repoPath, opts, fStats)

	if err != nil {
		return nil, fmt.Errorf("parse history failed: %w", err)
	}




	result.FileFunctionAnalyses = make(map[string]*models.FileAnalysis)
	cutoffDate := time.Now().AddDate(0, 0, -opts.TimeRangeDays)

	fileAnalyzer := NewFileAnalyzer(repoPath)

	

	for filePath := range result.FileStats {
		// Skip non-source files (docs, configs, etc.)
		if !isSourceFile(filePath) {
			continue
		}

		analysis, err := fileAnalyzer.AnalyzeFile(ctx, repo, filePath, cutoffDate)
		if err != nil {
			fmt.Println(err)
			continue
		}
		
		result.FileFunctionAnalyses[filePath] = analysis
	}

	return result, nil
}

func (ga *GitAnalyzer) parseGitHistory(ctx context.Context, repo *git.Repository, repoURL string, opts models.AnalyzeOptions, baseTreeStats map[string]bool) (*models.AnalysisResult, error) {
	branch := opts.Branch
	if branch == "" {
		branch = "main"
	}

	ref, err:= repo.Reference(plumbing.NewBranchReferenceName(branch), true)
	if err != nil {
		ref, err = repo.Reference(plumbing.NewTagReferenceName("master"), true)
		if err != nil {
			return nil, fmt.Errorf("get branch or tag failed: %w", err)
		}

		branch = "master"
	}

	// Get commit iterator
	commitIter, err := repo.Log(&git.LogOptions{
		From: ref.Hash(),
	})
	if err != nil {
		return nil, err
	}

	cutoffDate := time.Now().AddDate(0, 0, -opts.TimeRangeDays)

	fileStats:= make(map[string]*models.FileChangeStats)
	commitCount := 0

	err = commitIter.ForEach(func(c *object.Commit) error {

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if c.Author.When.Before(cutoffDate) {
			return nil
		}

		commitCount++

		stats, err := c.Stats()
		if err != nil {
			return err
		}

		for _, gitFileStat := range stats {
			path:= gitFileStat.Name


			if _, exists := fileStats[path]; !exists {
				if strings.Contains(path, "=>") {
					path = strings.TrimSpace(strings.Split(path, "=>")[1])
				}	
				fileStats[path] = &models.FileChangeStats{
					FilePath:          path,
					ChangesByDay:      make(map[int]int),
					UniqueAuthors:     make(map[string]int),
					FirstSeen:         c.Author.When,
				}
			}

			fs := fileStats[path]

			fs.TotalChanges ++

			if c.Author.When.After(fs.LastModified) {
				fs.LastModified = c.Author.When
			}

			dayOffset:= int(time.Since(c.Author.When).Hours() / 24)
			fs.ChangesByDay[dayOffset]++

			fs.UniqueAuthors[c.Author.Name]++
		}



		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("iterate commits failed: %w", err)
	}
	println("Commits processed: ", commitCount)
	println("Files processed: ", len(fileStats))
	println("file stats", fileStats)

	//filter stats to current tree
	for filePath := range fileStats {
		if _, exists := baseTreeStats[filePath]; !exists {
			println("Deleting file: ", filePath)
			delete(fileStats, filePath)
		}
	}

	return &models.AnalysisResult{
		RepoURL:      repoURL,
		Branch:       branch,
		AnalyzedAt:   time.Now(),
		CommitCount:  commitCount,
		FileStats:    fileStats,
		TimeRangeDays: opts.TimeRangeDays,
	}, nil
}

func isSourceFile(path string) bool {
	// Skip common non-source files
	excludePatterns := []string{
		".md", ".txt", ".json", ".yaml", ".yml",
		".xml", ".html", ".css", ".svg", ".png",
		".jpg", ".gif", ".pdf", ".lock", ".sum",
	}
	
	for _, pattern := range excludePatterns {
		if strings.HasSuffix(strings.ToLower(path), pattern) {
			return false
		}
	}
	
	return true
}