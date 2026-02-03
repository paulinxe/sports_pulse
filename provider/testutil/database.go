package testutil

import (
	"database/sql"
	"log/slog"
	"os"
	"provider/config"
	"testing"
)

func InitDB(t *testing.T) {
	err := config.InitDB()
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Verify database connection is ready
	if config.DB == nil {
		t.Fatalf("Database connection is nil after initialization")
	}

	_, _ = config.DB.Exec("TRUNCATE TABLE matches")
	_, _ = config.DB.Exec("TRUNCATE TABLE sync_state")
	_, _ = config.DB.Exec("TRUNCATE TABLE match_reconciliation")
}

func CloseDB() {
	if err := config.CloseDB(); err != nil {
		slog.Error("Failed to close database", "error", err)
		os.Exit(1)
	}
}

func BeginTransaction(t *testing.T) (*sql.Tx, error) {
	tx, err := config.DB.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	return tx, nil
}

func RollbackTransaction(t *testing.T, transaction *sql.Tx) {
	err := transaction.Rollback()
	if err != nil {
		t.Fatalf("Failed to rollback transaction: %v", err)
	}
}

func MatchExists(t *testing.T, canonicalID string) bool {
	var count int
	err := config.DB.QueryRow("SELECT COUNT(*) FROM matches WHERE canonical_id = $1", canonicalID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query database: %v", err)
	}
	return count > 0
}

func ReconciliationEntryExists(t *testing.T, providerMatchID string, provider int) bool {
	var count int
	err := config.DB.QueryRow("SELECT COUNT(*) FROM match_reconciliation WHERE provider_match_id = $1 AND provider = $2", providerMatchID, provider).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query database: %v", err)
	}
	return count > 0
}
