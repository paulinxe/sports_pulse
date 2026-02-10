package main

import (
	"log/slog"
	"os"
	"provider/internal/config"
	"provider/internal/repository"
	"provider/internal/service"
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

	args := os.Args[1:]
	if len(args) == 0 {
		slog.Error("Usage: <provider> <competition>  OR  --reconcile")
		slog.Error("Example: football_org la_liga")
		slog.Error("Example: --reconcile")
		os.Exit(int(BAD_ARGUMENTS))
	}

	db, err := config.InitDB()
	if err != nil {
		slog.Error("Failed to initialize database", "error", err)
		os.Exit(int(DB_INIT_ERROR))
	}
	defer func() { _ = db.Close() }()

	repositories, err := repository.InitRepositories(db)
	if err != nil {
		slog.Error("Failed to initialize repositories", "error", err)
		os.Exit(int(DB_INIT_ERROR))
	}
	clock := service.SystemClock{}

	if args[0] == "--reconcile" {
		if len(args) != 1 {
			slog.Error("Usage: --reconcile (no additional arguments)")
			os.Exit(int(BAD_ARGUMENTS))
		}

		slog.Debug("Starting reconciliation")
		if err := service.Reconcile(repositories); err != nil {
			slog.Error("Failed to reconcile", "error", err)
			os.Exit(int(PROVIDER_ERROR))
		}

		slog.Info("Reconciliation completed successfully")
		os.Exit(int(SUCCESS))
	}

	if len(args) != 2 {
		slog.Error("Usage: <provider> <competition>")
		slog.Error("Example: football_org la_liga")
		os.Exit(int(BAD_ARGUMENTS))
	}

	provider := args[0]
	competition := args[1]
	slog.Debug("Starting sync", "provider", strings.ToUpper(provider), "competition", strings.ToUpper(competition))

	if err := service.Sync(repositories, provider, competition, clock); err != nil {
		slog.Error("Failed to sync", "provider", provider, "competition", competition, "error", err)
		os.Exit(int(PROVIDER_ERROR))
	}

	slog.Info("Sync completed successfully", "provider", provider, "competition", competition)
	os.Exit(int(SUCCESS))
}
