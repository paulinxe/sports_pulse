package main

import (
	"log/slog"
	"os"
	"provider/config"
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

	if err := config.InitDB(); err != nil {
		slog.Error("Failed to initialize database", "error", err)
		os.Exit(int(DB_INIT_ERROR))
	}
	defer func() { _ = config.CloseDB() }()

	provider := os.Args[1]
	competition := os.Args[2]
	slog.Debug("Starting sync", "provider", strings.ToUpper(provider), "competition", strings.ToUpper(competition))

	if err := Sync(provider, competition, systemClock{}); err != nil {
		slog.Error("Failed to sync", "provider", provider, "competition", competition, "error", err)
		os.Exit(int(PROVIDER_ERROR))
	}

	slog.Info("Sync completed successfully", "provider", provider, "competition", competition)
	os.Exit(int(SUCCESS))
}
