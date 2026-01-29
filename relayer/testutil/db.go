package testutil

import (
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"relayer/config"
)

func InitDatabase(t *testing.T) {
	t.Helper()
	_, err := config.InitDB()
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Verify database connection is ready
	if config.DB == nil {
		t.Fatalf("Database connection is nil after initialization")
	}

	_, _ = config.DB.Exec("TRUNCATE TABLE matches")
}

const signedStatus = 4

func InsertSignedMatch(t *testing.T, id uuid.UUID, canonicalID string, competitionID, homeTeamID, awayTeamID, homeTeamScore, awayTeamScore int32, start time.Time, signatureHex string) {
	t.Helper()
	if config.DB == nil {
		t.Fatal("database not initialized")
	}

	_, err := config.DB.Exec(`INSERT INTO matches (id, canonical_id, competition_id, home_team_id, away_team_id, home_team_score, away_team_score, start, "end", signature, status, provider_match_id, provider)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, 'dummy-provider-match-id', 1)`,
		id, canonicalID, competitionID, homeTeamID, awayTeamID, homeTeamScore, awayTeamScore, start, start, signatureHex, signedStatus)

	if err != nil {
		t.Fatalf("insert signed match: %v", err)
	}
}

func CloseDatabase() {
	if err := config.Close(); err != nil {
		slog.Error("Failed to close database", "error", err)
		os.Exit(1)
	}
}
