package testutil

import (
	"database/sql"
	"log/slog"
	"os"
	"signer/db"
	"testing"
)

func InitDatabase(t *testing.T) {
	_, err := db.Init()
	if err != nil {
		t.Fatalf("failed to initialize database: %v", err)
	}

	// Verify database connection is ready
	if db.DB == nil {
		t.Fatalf("Database connection is nil after initialization")
	}

	_, _ = db.DB.Exec("TRUNCATE TABLE matches")
}

func CloseDatabase() {
	if err := db.Close(); err != nil {
		slog.Error("failed to close database", "error", err)
		os.Exit(1)
	}
}

func BeginTransaction(t *testing.T) (*sql.Tx, error) {
	tx, err := db.DB.Begin()
	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}
	return tx, nil
}

func RollbackTransaction(t *testing.T, transaction *sql.Tx) {
	err := transaction.Rollback()
	if err != nil {
		t.Fatalf("failed to rollback transaction: %v", err)
	}
}

func MatchExists(t *testing.T, canonicalID string) bool {
	var count int
	err := db.DB.QueryRow("SELECT COUNT(*) FROM matches WHERE canonical_id = $1", canonicalID).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query database: %v", err)
	}
	return count > 0
}
