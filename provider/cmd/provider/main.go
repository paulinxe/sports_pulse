package main

import (
	"log/slog"
	"os"
	"provider/internal/config"
	"provider/internal/repository"
	"provider/internal/sync"
	"strings"
)

type ErrorCodes int

const (
	SUCCESS ErrorCodes = iota
	BAD_ARGUMENTS
	DB_INIT_ERROR
	PROVIDER_ERROR
)

func main() {
	slog.SetDefault(
		slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})),
	)

	if len(os.Args) != 3 {
		slog.Error("Usage: <provider> <competition>")
		slog.Error("Example: football_org la_liga")
		os.Exit(int(BAD_ARGUMENTS))
	}

	db, err := config.InitDB()
	if err != nil {
		slog.Error("Failed to initialize database", "error", err)
		os.Exit(int(DB_INIT_ERROR))
	}
	defer func() { _ = db.Close() }()

	repositories := repository.InitRepositories(db)

	provider := os.Args[1]
	competition := os.Args[2]
	slog.Debug("Starting sync", "provider", strings.ToUpper(provider), "competition", strings.ToUpper(competition))

	if err := sync.Sync(repositories, provider, competition, sync.SystemClock{}); err != nil {
		slog.Error("Failed to sync", "provider", provider, "competition", competition, "error", err)
		os.Exit(int(PROVIDER_ERROR))
	}

	slog.Info("Sync completed successfully", "provider", provider, "competition", competition)
	os.Exit(int(SUCCESS))
}
