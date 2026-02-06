package testutil

import (
	"context"
	"database/sql"
	"signer/internal/config"
	"testing"
)

// InitDB initializes the database, truncates matches, and returns the connection.
// The caller must call db.Close() when done (e.g. defer db.Close()).
func InitDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := config.InitDB()
	if err != nil {
		t.Fatalf("failed to initialize database: %v", err)
	}
	_, _ = db.Exec("TRUNCATE TABLE matches")
	return db
}

func BeginTransaction(t *testing.T, ctx context.Context, db *sql.DB) (*sql.Tx, error) {
	t.Helper()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}
	return tx, nil
}

func RollbackTransaction(t *testing.T, tx *sql.Tx) {
	t.Helper()
	if err := tx.Rollback(); err != nil {
		t.Fatalf("failed to rollback transaction: %v", err)
	}
}

func MatchExists(t *testing.T, ctx context.Context, db *sql.DB, canonicalID string) bool {
	t.Helper()
	var count int
	err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM matches WHERE canonical_id = $1", canonicalID).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query database: %v", err)
	}
	return count > 0
}
