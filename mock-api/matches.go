package main

import (
	"math/rand"
	"mock_api/repository"
)

const (
	LaLigaCompetitionID = 2014
	StatusFinished      = "FINISHED"
)

func CreateMatch(scheduled ScheduledMatch) repository.Match {
	homeScore, awayScore := generateScores()
	
	return repository.Match{
		CompetitionID:  LaLigaCompetitionID,
		UTCDate:        scheduled.Date.Format("2006-01-02T15:04:05Z"),
		Status:         StatusFinished,
		HomeTeamID:     scheduled.HomeTeamID,
		AwayTeamID:     scheduled.AwayTeamID,
		HomeTeamScore:  homeScore,
		AwayTeamScore:  awayScore,
		Matchday:       scheduled.Matchday,
	}
}

// GenerateScores generates random but realistic scores for a match
// Scores range from 0-5, weighted towards 0-3 goals
func generateScores() (homeScore, awayScore uint) {
	homeScore = generateWeightedScore()
	awayScore = generateWeightedScore()
	return homeScore, awayScore
}

// generateWeightedScore generates a score with weighted distribution
// 0-2: 70% (0: 30%, 1: 25%, 2: 15%)
// 3: 25%
// 4-5: 5% (4: 3%, 5: 2%)
func generateWeightedScore() uint {
	r := rand.Float64()
	
	if r < 0.30 {
		return 0
	}
	
	if r < 0.55 {
		return 1
	}
	
	if r < 0.70 {
		return 2
	}
	
	if r < 0.95 {
		return 3
	}
	
	if r < 0.98 {
		return 4
	}
	
	return 5
}