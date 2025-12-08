package testutil

import (
    "provider/db"
    "log/slog"
    "os"
    "testing"
)

func InitDatabase(t *testing.T) {
    err := db.Init()
    if err != nil {
        t.Fatalf("Failed to initialize database: %v", err)
    }

    // Verify database connection is ready
    if db.DB == nil {
        t.Fatalf("Database connection is nil after initialization")
    }

    // Clean up before the test
    // TODO: we need a better way to clean up the database.
    _, _ = db.DB.Exec("DELETE FROM matches")
}

func CloseDatabase() {
    if err := db.Close(); err != nil {
        slog.Error("Failed to close database", "error", err)
        os.Exit(1)
    }
}

func MatchExists(t *testing.T, canonicalID string) bool {
    var count int
    err := db.DB.QueryRow("SELECT COUNT(*) FROM matches WHERE canonical_id = $1", canonicalID).Scan(&count)
    if err != nil {
        t.Fatalf("Failed to query database: %v", err)
    }
    return count > 0
}