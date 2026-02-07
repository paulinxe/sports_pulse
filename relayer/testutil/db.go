package testutil

import (
	"database/sql"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"relayer/internal/config"
	"relayer/internal/repository"

	"github.com/google/uuid"
)

func InitDB(t *testing.T) (*sql.DB, *repository.MatchRepository) {
	t.Helper()
	db, _, err := config.InitDB()
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	_, _ = db.Exec("TRUNCATE TABLE matches")
	repo := repository.NewMatchRepository(db)
	return db, repo
}

const signedStatus = 4

func InsertSignedMatch(t *testing.T, db *sql.DB, id uuid.UUID, canonicalID string, competitionID, homeTeamID, awayTeamID, homeTeamScore, awayTeamScore int32, start time.Time, signatureHex string) {
	t.Helper()
	if db == nil {
		t.Fatal("database not initialized")
	}

	_, err := db.Exec(`INSERT INTO matches (id, canonical_id, competition_id, home_team_id, away_team_id, home_team_score, away_team_score, start, "end", signature, status, provider_match_id, provider)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, 'dummy-provider-match-id', 1)`,
		id, canonicalID, competitionID, homeTeamID, awayTeamID, homeTeamScore, awayTeamScore, start, start, signatureHex, signedStatus)

	if err != nil {
		t.Fatalf("insert signed match: %v", err)
	}
}

func CloseDB(db *sql.DB) {
	if db == nil {
		return
	}
	if err := db.Close(); err != nil {
		slog.Error("Failed to close database", "error", err)
		os.Exit(1)
	}
}

func QueryMatchStatus(db *sql.DB, matchID uuid.UUID, dest *int) error {
	if db == nil {
		return errors.New("database not initialized")
	}
	return db.QueryRow("SELECT status FROM matches WHERE id = $1", matchID).Scan(dest)
}
