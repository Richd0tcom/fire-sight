package analyzer

import (
	"context"
	"fmt"
	"os"
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

func (ga *GitAnalyzer) AnalyzeRepository(ctx context.Context, repoPath string, opts models.AnalyzeOptions) (*models.AnalysisResult, error) {

	repo, _, err := ga.cloneRepo(ctx, repoPath, opts)
	if err != nil {
		return nil, fmt.Errorf("clone failed: %w", err)
	}

	defer os.RemoveAll(repoPath) 
	
	result, err := ga.parseGitHistory(ctx, repo, repoPath, opts)

	if err != nil {
		return nil, fmt.Errorf("parse history failed: %w", err)
	}

	return result, nil
}

func (ga *GitAnalyzer) parseGitHistory(ctx context.Context, repo *git.Repository, repoURL string, opts models.AnalyzeOptions) (*models.AnalysisResult, error) {
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

	return &models.AnalysisResult{
		RepoURL:      repoURL,
		Branch:       branch,
		AnalyzedAt:   time.Now(),
		CommitCount:  commitCount,
		FileStats:    fileStats,
		TimeRangeDays: opts.TimeRangeDays,
	}, nil
}