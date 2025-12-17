package repository

import (
	"database/sql"
	"fmt"
	"log/slog"
	"mock_api/db"
	"time"
)

type Match struct {
	ID            uint
	CompetitionID uint
	UTCDate       string
	Status        string
	HomeTeamID    uint
	AwayTeamID    uint
	HomeTeamScore uint
	AwayTeamScore uint
	Matchday      int
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

func FindMatchesByDateRange(competitionID int, dateFrom, dateTo time.Time) ([]Match, error) {
	if db.DB == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	query := `
		SELECT id, competition_id, utc_date, status, home_team_id, away_team_id,
		       home_team_score, away_team_score, matchday
		FROM test_football_org_matches
		WHERE competition_id = $1 
		  AND DATE(utc_date) >= DATE($2) 
		  AND DATE(utc_date) <= DATE($3)
		ORDER BY utc_date ASC
	`

	rows, err := db.DB.Query(query, competitionID, dateFrom, dateTo)
	if err != nil {
		return nil, fmt.Errorf("failed to query matches: %w", err)
	}
	defer rows.Close()

	var matches []Match
	for rows.Next() {
		var match Match
		var utcDate time.Time
		err := rows.Scan(
			&match.ID,
			&match.CompetitionID,
			&utcDate,
			&match.Status,
			&match.HomeTeamID,
			&match.AwayTeamID,
			&match.HomeTeamScore,
			&match.AwayTeamScore,
			&match.Matchday,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan match: %w", err)
		}
		match.UTCDate = utcDate.Format("2006-01-02T15:04:05Z")
		matches = append(matches, match)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating matches: %w", err)
	}

	return matches, nil
}

func FindMatchByID(matchID int) (*Match, error) {
	if db.DB == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	query := `
		SELECT id, competition_id, utc_date, status, home_team_id, away_team_id,
		       home_team_score, away_team_score, matchday
		FROM test_football_org_matches
		WHERE id = $1
	`

	var match Match
	var utcDate time.Time
	err := db.DB.QueryRow(query, matchID).Scan(
		&match.ID,
		&match.CompetitionID,
		&utcDate,
		&match.Status,
		&match.HomeTeamID,
		&match.AwayTeamID,
		&match.HomeTeamScore,
		&match.AwayTeamScore,
		&match.Matchday,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("match not found: %d", matchID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query match: %w", err)
	}

	match.UTCDate = utcDate.Format("2006-01-02T15:04:05Z")
	return &match, nil
}

