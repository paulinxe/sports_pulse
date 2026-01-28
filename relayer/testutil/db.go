package testutil

import (
	"log/slog"
	"os"
	"relayer/config"
	"testing"
)

func InitDatabase(t *testing.T) {
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

func CloseDatabase() {
	if err := config.Close(); err != nil {
		slog.Error("Failed to close database", "error", err)
		os.Exit(1)
	}
}
