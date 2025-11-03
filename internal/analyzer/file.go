package analyzer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/richd0tcom/fire-sight/internal/models"
	"github.com/richd0tcom/fire-sight/internal/parser"
)

// LEARNING MOMENT: Mapping Git Changes to Functions
//
// Algorithm:
// 1. Parse file → Extract functions with line ranges
// 2. Build FunctionMap (sorted intervals) for O(log n) lookup
// 3. For each git commit:
//    - Get changed lines (git blame / diff)
//    - Binary search: line → function
//    - Increment function's change count
// 4. Calculate heat scores per function

type FileAnalyzer struct {
	repoPath string
}

func NewFileAnalyzer(repoPath string) *FileAnalyzer {
	return &FileAnalyzer{repoPath: repoPath}
}

// AnalyzeFile reads file content and extracts function-level metrics
func (fa *FileAnalyzer) AnalyzeFile(
	ctx context.Context,
	repo *git.Repository,
	filePath string,
	fileStats *models.FileChangeStats,
	cutoffDate time.Time,
) (*models.FileAnalysis, error) {
	// Read file content
	fullPath := filepath.Join(fa.repoPath, filePath)
	file, err := os.Open(fullPath)
	if err != nil {
		// File might have been deleted - return empty analysis
		return &models.FileAnalysis{
			Path:      filePath,
			Functions: []*models.FunctionStats{},
		}, nil
	}
	defer file.Close()

	// Detect language
	lang := parser.DetectLanguage(filePath)
	if !parser.IsSupported(lang) {
		// Unsupported language - return file-level stats only
		return &models.FileAnalysis{
			Path:      filePath,
			Language:  string(lang),
			Functions: []*models.FunctionStats{},
		}, nil
	}

	// Parse functions
	p := parser.GetParser(lang)
	functions, err := p.Parse(file)
	if err != nil {
		return nil, fmt.Errorf("parse failed: %w", err)
	}

	// Build function map for fast lookups
	fnMap := parser.NewFunctionMap(functions)

	// Map git changes to functions
	functionStats := fa.mapChangesToFunctions(repo, filePath, fnMap, cutoffDate)

	return &models.FileAnalysis{
		Path:      filePath,
		Language:  string(lang),
		Functions: functionStats,
	}, nil
}

// mapChangesToFunctions uses git log to find which lines changed, then maps to functions
func (fa *FileAnalyzer) mapChangesToFunctions(
	repo *git.Repository,
	filePath string,
	fnMap *parser.FunctionMap,
	cutoffDate time.Time,
) []*models.FunctionStats {
	// Initialize stats for each function
	statsMap := make(map[string]*models.FunctionStats)
	for _, fn := range fnMap.GetAll() {
		statsMap[fn.Name] = &models.FunctionStats{
			Name:         fn.Name,
			LineStart:    fn.LineStart,
			LineEnd:      fn.LineEnd,
			ChangesByDay: make(map[int]int),
			LastModified: time.Time{},
		}
	}

	// Get file history
	commits, err := repo.Log(&git.LogOptions{
		FileName: &filePath,
	})
	if err != nil {
		return fa.statsToSlice(statsMap)
	}

	// Process each commit
	commits.ForEach(func(c *object.Commit) error {
		if c.Author.When.Before(cutoffDate) {
			return nil // Skip old commits
		}

		// Get patch (changed lines)
		patch, err := fa.getCommitPatch(repo, c, filePath)
		if err != nil {
			return nil // Skip on error
		}

		// Extract changed line numbers from patch
		changedLines := fa.extractChangedLines(patch)

		// Map each changed line to its function
		for _, line := range changedLines {
			fn := fnMap.FindByLine(line)
			if fn != nil {
				stats := statsMap[fn.Name]
				stats.TotalChanges++
				
				// Update last modified
				if c.Author.When.After(stats.LastModified) {
					stats.LastModified = c.Author.When
				}

				// Track by day offset
				dayOffset := int(time.Since(c.Author.When).Hours() / 24)
				stats.ChangesByDay[dayOffset]++
			}
		}

		return nil
	})

	return fa.statsToSlice(statsMap)
}

// getCommitPatch gets the diff for a specific file in a commit
func (fa *FileAnalyzer) getCommitPatch(repo *git.Repository, commit *object.Commit, filePath string) (string, error) {
	// Get parent commit
	parent, err := commit.Parent(0)
	if err != nil {
		// No parent (first commit) - all lines are "new"
		return "", nil
	}

	// Get patch between parent and current
	patch, err := parent.Patch(commit)
	if err != nil {
		return "", err
	}

	return patch.String(), nil
}

// extractChangedLines parses git diff patch format to get line numbers
func (fa *FileAnalyzer) extractChangedLines(patch string) []int {
	lines := []int{}
	currentLine := 0

	for _, line := range strings.Split(patch, "\n") {
		if strings.HasPrefix(line, "@@") {
			// Parse hunk header: @@ -10,5 +12,7 @@
			// The "+12,7" means starting at line 12 in new file
			parts := strings.Split(line, " ")
			for _, part := range parts {
				if strings.HasPrefix(part, "+") {
					// Parse "+12,7" or "+12"
					numStr := strings.TrimPrefix(part, "+")
					numStr = strings.Split(numStr, ",")[0]
					fmt.Sscanf(numStr, "%d", &currentLine)
					break
				}
			}
		} else if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			// Added or modified line
			lines = append(lines, currentLine)
			currentLine++
		} else if !strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "\\") {
			// Context line (unchanged)
			currentLine++
		}
	}

	return lines
}

// statsToSlice converts map to slice
func (fa *FileAnalyzer) statsToSlice(statsMap map[string]*models.FunctionStats) []*models.FunctionStats {
	result := make([]*models.FunctionStats, 0, len(statsMap))
	for _, stats := range statsMap {
		result = append(result, stats)
	}
	return result
}