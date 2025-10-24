package analyzer

import (
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/richd0tcom/fire-sight/internal/models"
)


type TreeBuilder struct {
	// Maps path -> node for O(1) lookups during construction
	nodeCache map[string]*models.FileNode
}

func NewTreeBuilder() *TreeBuilder {
	return &TreeBuilder{
		nodeCache: make(map[string]*models.FileNode),
	}
}

func (tb *TreeBuilder) BuildTree(heatScores []models.HeatScore, fileStats map[string]*models.FileChangeStats) *models.FileNode {
	// Create virtual root node
	root := &models.FileNode{
		ID:       "root",
		Name:     "root",
		Path:     "",
		Type:     "folder",
		Children: []*models.FileNode{},
	}
	tb.nodeCache[""] = root

	// Process each file
	for _, score := range heatScores {
		tb.addFileToTree(root, score, fileStats[score.Path])
	}

	// Calculate aggregated stats for folders (bottom-up)
	tb.aggregateFolderStats(root)

	return root
}

// addFileToTree inserts a file into the tree, creating parent folders as needed
func (tb *TreeBuilder) addFileToTree(root *models.FileNode, score models.HeatScore, stats *models.FileChangeStats) {
	// Split path into parts: "src/components/Button.tsx" -> ["src", "components", "Button.tsx"]
	parts := strings.Split(score.Path, "/")
	
	currentPath := ""
	currentNode := root

	// Walk/create path to file
	for i, part := range parts {
		if part == "" {
			continue
		}

		// Build cumulative path
		if currentPath == "" {
			currentPath = part
		} else {
			currentPath = currentPath + "/" + part
		}

		isFile := i == len(parts)-1

		// Check if node already exists (O(1) lookup)
		node, exists := tb.nodeCache[currentPath]
		
		if !exists {
			// Create new node
			node = &models.FileNode{
				ID:       generateNodeID(currentPath),
				Name:     part,
				Path:     currentPath,
				Type:     tb.getNodeType(isFile),
				Children: []*models.FileNode{},
			}

			// Add file-specific data
			if isFile {
				node.Extension = tb.getExtension(part)
				node.Size = 0 // TODO: Add in Milestone 2 when we parse files
				node.LinesOfCode = 0
				node.LastModified = stats.LastModified
		
				node.HeatScore = &score
				node.Functions = []*models.FunctionNode{} // Populated in Milestone 2
			}

			// Add to parent's children
			currentNode.Children = append(currentNode.Children, node)
			
			// Cache for O(1) future lookups
			tb.nodeCache[currentPath] = node
		}

		currentNode = node
	}
}

// aggregateFolderStats calculates folder metrics from children (recursive, bottom-up)
func (tb *TreeBuilder) aggregateFolderStats(node *models.FileNode) {
	if node.Type == "file" {
		return // Base case: files already have stats
	}

	// Recursively process children first
	for _, child := range node.Children {
		tb.aggregateFolderStats(child)
	}

	// Aggregate from children
	var (
		totalFiles      int
		totalSize       int64
		totalLines      int
		totalChanges    int
		latestModified  time.Time
		weightedHeat    float64
		totalHeatWeight int
	)

	for _, child := range node.Children {
		if child.Type == "file" {
			totalFiles++
			totalSize += child.Size
			totalLines += child.LinesOfCode
			totalChanges += child.HeatScore.TotalFileChanges
			
			// Track latest modification
			if child.LastModified.After(latestModified) {
				latestModified = child.LastModified
			}

			// Weighted average for heat score
			// Weight by change frequency (active files matter more)
			weight := child.HeatScore.TotalFileChanges
			if weight == 0 {
				weight = 1 
			}
			weightedHeat += child.HeatScore.Score * float64(weight)
			totalHeatWeight += weight
		} else {
			// Folder - aggregate its aggregated stats
			if child.HeatScore == nil {
				child.HeatScore = &models.HeatScore{
					Path: child.Path,
					Score: 0,
					ChangeFreq: 0,
					DaysSinceEdit: 0,
					TotalFileChanges: 0,
				}
			}
			totalFiles += child.FileCount
			totalSize += child.Size
			totalLines += child.LinesOfCode
			totalChanges += child.HeatScore.TotalFileChanges
			
			if child.LastModified.After(latestModified) {
				latestModified = child.LastModified
			}

			weight := child.HeatScore.TotalFileChanges
			if weight == 0 {
				weight = 1
			}
			
			weightedHeat += child.HeatScore.Score * float64(weight)
			totalHeatWeight += weight
		}
	}

	// Set aggregated values
	node.FileCount = totalFiles
	node.Size = totalSize
	node.LinesOfCode = totalLines
	node.LastModified = latestModified

	if node.HeatScore == nil {
		node.HeatScore = &models.HeatScore{
			Path: node.Path,
			Score: 0,
			ChangeFreq: 0,
			DaysSinceEdit: 0,
			TotalFileChanges: 0,
		}
	}

	node.HeatScore.TotalFileChanges = totalChanges
	
	// Calculate weighted average heat score
	if totalHeatWeight > 0 {
		node.HeatScore.Score = weightedHeat / float64(totalHeatWeight)
	}

	// Sort children: folders first (alphabetically), then files (by heat score desc)
	tb.sortChildren(node)
}

// sortChildren orders children for optimal UI display
func (tb *TreeBuilder) sortChildren(node *models.FileNode) {
	sort.SliceStable(node.Children, func(i, j int) bool {
		a, b := node.Children[i], node.Children[j]

		// Folders before files
		if a.Type == "folder" && b.Type == "file" {
			return true
		}
		if a.Type == "file" && b.Type == "folder" {
			return false
		}

		// Within same type:
		if a.Type == "folder" {
			// Folders: alphabetical
			return a.Name < b.Name
		} else {
			// Files: by heat score (hottest first)
			return a.HeatScore.Score > b.HeatScore.Score
		}
	})
}

// Helper functions
func (tb *TreeBuilder) getNodeType(isFile bool) models.FileNodeType {
	if isFile {
		return "file"
	}
	return "folder"
}

func (tb *TreeBuilder) getExtension(filename string) string {
	ext := filepath.Ext(filename)
	if ext != "" {
		return ext[1:] // Remove leading dot
	}
	return ""
}

func generateNodeID(path string) string {
	// Simple ID generation - could use UUID for uniqueness
	return strings.ReplaceAll(path, "/", "_")
}
