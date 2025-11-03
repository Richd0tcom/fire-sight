package analyzer

import (
	"math"
	"time"

	"github.com/richd0tcom/fire-sight/internal/models"
)


const (
	DecayRate = 0.01
	MinScore = 1.0
)

type HeatCalculator struct {}

func NewHeatCalculator() *HeatCalculator {
	return &HeatCalculator{}
}

func (hc *HeatCalculator) CalculateHeatScores(result *models.AnalysisResult) []models.HeatScore {
	scores := make([]models.HeatScore, 0, len(result.FileStats))

	now:= time.Now()

	// First pass: calculate raw scores
	maxRawScore := 0.0
	rawScores := make(map[string]float64)

	for path, fs := range result.FileStats {
		rawScore := hc.calculateRawScore(fs)
		rawScores[path] = rawScore
		
		if rawScore > maxRawScore {
			maxRawScore = rawScore
		}
	}


	for path, stats := range result.FileStats {
		rawScore := rawScores[path]
		
		// Normalize: (raw / max) * 100
		// This ensures hottest file = 100, coldest = relative to that
		normalizedScore := MinScore
		if maxRawScore > 0 {
			normalizedScore = (rawScore / maxRawScore) * 100
		}

		daysSinceEdit := int(now.Sub(stats.LastModified).Hours() / 24)

		scores = append(scores, models.HeatScore{
			Path:          path,
			Score:         normalizedScore,
			ChangeFreq:    hc.calculateChangeFrequency(stats.TotalChanges, daysSinceEdit),
			DaysSinceEdit: daysSinceEdit,
			TotalFileChanges: stats.TotalChanges,
		})
	}

	

	return scores
}

func (hc *HeatCalculator) calculateRawScore(fs *models.FileChangeStats) float64 {
	score:= 0.0

	for dayOffset, changeCount := range fs.ChangesByDay {
		weight:= getTimeDecay(dayOffset)

		score += float64(changeCount) * weight
	}

	authorBonus:= getAuthorBonus(len(fs.UniqueAuthors))
	score *= (1.0 + authorBonus)

	return score
}

//returns num of changes per week
func (hc *HeatCalculator) calculateChangeFrequency(totalChanges int, days int) float64 {
	if days == 0 {
		return 0
	}
	
	weeks := float64(days) / 7.0
	return float64(totalChanges) / weeks
}



func (hc *HeatCalculator) isLikelyDeadCode(stats *models.FileChangeStats, now time.Time) bool {
	// Signal 1: No changes in 6+ months
	sixMonthsAgo := now.AddDate(0, -6, 0)
	noRecentChanges := stats.LastModified.Before(sixMonthsAgo)

	// Signal 2: Very few total changes (< 3)
	fewChanges := stats.TotalChanges < 3

	// Signal 3: Single author (might be abandoned experiment)
	singleAuthor := len(stats.UniqueAuthors) == 1

	// Dead code if 2 out of 3 signals present
	signalCount := 0
	if noRecentChanges {

		signalCount++
	}
	if fewChanges {
		signalCount++
	}
	if singleAuthor {
		signalCount++
	}

	return signalCount >= 2
}

func getTimeDecay(daysSinceEdit int) float64 {
	return math.Exp(-DecayRate * float64(daysSinceEdit))
}

func getAuthorBonus(numUniqueAuthors int) float64 {
	return math.Min(float64(numUniqueAuthors)*0.1, 0.5)
}


func (hc *HeatCalculator) calculateRawFunctionScore(stats *models.FunctionStats) float64 {
	score := 0.0

	for dayOffset, changeCount := range stats.ChangesByDay {
		weight:= getTimeDecay(dayOffset)

		score += float64(changeCount) * weight
	}

	return score
}

// IsLikelyDeadFunctionCode determines if a function is probably unused
func (hc *HeatCalculator) isLikelyDeadFunctionCode(stats *models.FunctionStats, now time.Time) bool {
	// Signal 1: No changes in 6+ months
	sixMonthsAgo := now.AddDate(0, -6, 0)
	noRecentChanges := stats.LastModified.Before(sixMonthsAgo)

	// Signal 2: Very few total changes (< 2)
	fewChanges := stats.TotalChanges < 2

	// Dead code if both signals present
	return noRecentChanges && fewChanges
}

func (hc *HeatCalculator) CalculateFunctionHeatScore(stats []*models.FunctionStats) []*models.FunctionNode {
	result := make([]*models.FunctionNode, 0, len(stats))

	maxRawScore := 0.0
	rawScores := make(map[string]float64)

	for _, stat := range stats {
		rawScore := hc.calculateRawFunctionScore(stat)

		rawScores[stat.Name] = rawScore
		
		if rawScore > maxRawScore {
			maxRawScore = rawScore
		}
	}


	//Normalize
	for _, stat := range stats {
		rawScore := rawScores[stat.Name]
		
		// Normalize: (raw / max) * 100
		// This ensures hottest file = 100, coldest = relative to that
		normalizedScore := MinScore
		if maxRawScore > 0 {
			normalizedScore = (rawScore / maxRawScore) * 100
		}

		result = append(result, &models.FunctionNode{
			Name:            stat.Name,
			LineStart:       stat.LineStart,
			LineEnd:         stat.LineEnd,
			LastModified:    stat.LastModified,
			HeatScore:       &models.HeatScore{
				Score:         normalizedScore,
				ChangeFreq:    hc.calculateChangeFrequency(stat.TotalChanges, int(time.Since(stat.LastModified).Hours()/24)),
				DaysSinceEdit: int(time.Since(stat.LastModified).Hours()/24),
			},
			IsDeadCode:      hc.isLikelyDeadFunctionCode(stat, time.Now()),
		})
	}

	return result
}