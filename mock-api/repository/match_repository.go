package repository

import (
	"fmt"
	"log/slog"
	"mock_api/db"
)

type Match struct {
	CompetitionID  uint
	UTCDate        string
	Status         string
	HomeTeamID     uint
	AwayTeamID     uint
	HomeTeamScore  uint
	AwayTeamScore  uint
	Matchday       int
}

func ClearAllMatches() error {
	if db.DB == nil {
		return fmt.Errorf("database connection not initialized")
	}

	query := `TRUNCATE TABLE test_football_org_matches`
	_, err := db.DB.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to clear matches: %w", err)
	}

	slog.Info("Cleared all matches from test_football_org_matches table")
	return nil
}

// InsertMatchesBatch inserts multiple matches in a transaction
func InsertMatchesBatch(matches []Match) error {
	if db.DB == nil {
		return fmt.Errorf("database connection not initialized")
	}

	tx, err := db.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO test_football_org_matches (
			competition_id, utc_date, status, home_team_id, away_team_id,
			home_team_score, away_team_score, matchday
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	stmt, err := tx.Prepare(query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, match := range matches {
		_, err := stmt.Exec(
			match.CompetitionID,
			match.UTCDate,
			match.Status,
			match.HomeTeamID,
			match.AwayTeamID,
			match.HomeTeamScore,
			match.AwayTeamScore,
			match.Matchday,
		)
		if err != nil {
			return fmt.Errorf("failed to insert match: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	slog.Info("Inserted matches", "count", len(matches))
	return nil
}

