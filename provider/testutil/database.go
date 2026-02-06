package testutil

import (
	"database/sql"
	"log/slog"
	"os"
	"provider/internal/config"
	"provider/internal/repository"
	"testing"
)

// InitDB initializes the database and returns it along with repositories.
// The caller should call db.Close() when done.
func InitDB(t *testing.T) (*sql.DB, *repository.Repositories) {
	t.Helper()
	db, err := config.InitDB()
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	_, _ = db.Exec("TRUNCATE TABLE matches")
	_, _ = db.Exec("TRUNCATE TABLE sync_state")
	_, _ = db.Exec("TRUNCATE TABLE match_reconciliation")

	return db, repository.InitRepositories(db)
}

func CloseDB(db *sql.DB) {
	if db != nil {
		if err := db.Close(); err != nil {
			slog.Error("Failed to close database", "error", err)
			os.Exit(1)
		}
	}
}

func BeginTransaction(t *testing.T, db *sql.DB) (*sql.Tx, error) {
	t.Helper()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	return tx, nil
}

func RollbackTransaction(t *testing.T, transaction *sql.Tx) {
	t.Helper()
	err := transaction.Rollback()
	if err != nil {
		t.Fatalf("Failed to rollback transaction: %v", err)
	}
}

func MatchExists(t *testing.T, db *sql.DB, canonicalID string) bool {
	t.Helper()
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM matches WHERE canonical_id = $1", canonicalID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query database: %v", err)
	}
	return count > 0
}

func ReconciliationEntryExists(t *testing.T, db *sql.DB, providerMatchID string, provider int) bool {
	t.Helper()
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM match_reconciliation WHERE provider_match_id = $1 AND provider = $2", providerMatchID, provider).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query database: %v", err)
	}
	return count > 0
}
