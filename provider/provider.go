package main

import (
	"log/slog"
	"os"
	"provider/db"
	"provider/entity"
	"provider/football_org"
	"strings"
)

type ErrorCodes int

const (
	SUCCESS ErrorCodes = iota
	BAD_ARGUMENTS
	UNKNOWN_PROVIDER
	UNKNOWN_COMPETITION
	DB_INIT_ERROR
	PROVIDER_ERROR
)

func main() {
	os.Exit(run(os.Args))
}

func run(args []string) int {
	slog.SetDefault(
		slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})),
	)

	if len(args) != 3 {
		slog.Error("Usage: <provider> <competition>")
		return int(BAD_ARGUMENTS)
	}

	competition := entity.Competition(0)
	switch strings.ToLower(args[2]) {
	case "la_liga":
		competition = entity.LaLiga
	default:
		slog.Error("Unknown competition", "competition", args[2])
		return int(UNKNOWN_COMPETITION)
	}

	if err := db.Init(); err != nil {
		slog.Error("Failed to initialize database", "error", err)
		return int(DB_INIT_ERROR)
	}
	defer db.Close()

	provider := strings.ToLower(args[1])
	slog.Debug("Initializing connection to provider", "provider", strings.ToUpper(provider))
	switch provider {
	case "football_org":
		if err := football_org.Sync(competition); err != nil {
			slog.Error("Failed to sync Football Data API", "error", err)
			return int(PROVIDER_ERROR)
		}
	default:
		slog.Error("Unknown provider", "provider", provider)
		return int(UNKNOWN_PROVIDER)
	}

	slog.Info("Operation completed successfully", "provider", provider)
	return int(SUCCESS)
}
