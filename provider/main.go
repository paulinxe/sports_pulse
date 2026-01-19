package main

import (
	"log/slog"
	"os"
	"provider/db"
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
		slog.Error("Example: football_org la_liga")
		return int(BAD_ARGUMENTS)
	}

	if err := db.Init(); err != nil {
		slog.Error("Failed to initialize database", "error", err)
		return int(DB_INIT_ERROR)
	}
	defer db.Close()

	provider := args[1]
	competition := args[2]
	slog.Debug("Starting sync", "provider", strings.ToUpper(provider), "competition", strings.ToUpper(competition))

	if err := Sync(provider, competition); err != nil {
		slog.Error("Failed to sync", "provider", provider, "competition", competition, "error", err)
		return int(PROVIDER_ERROR)
	}

	slog.Info("Sync completed successfully", "provider", provider, "competition", competition)
	return int(SUCCESS)
}
