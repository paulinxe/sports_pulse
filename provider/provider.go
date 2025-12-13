package main

import (
	"log/slog"
	"os"
	"provider/db"
	"provider/entity"
	"provider/football_org"
	"strings"
)

func main() {
	os.Exit(run(os.Args))
}

func run(args []string) int {
	// Parse arguments and check for --debug flag
	debugMode := false
	filteredArgs := []string{args[0]} // Keep program name

	for i := 1; i < len(args); i++ {
		if args[i] == "--debug" {
			debugMode = true
			continue
		}
		filteredArgs = append(filteredArgs, args[i])
	}

	if len(filteredArgs) != 3 {
		slog.Error("Usage: provider <provider> <competition> [--debug]")
		return 1
	}

	// Set up debug logging if flag is present
	if debugMode {
		slog.SetDefault(
			slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			})),
		)
	}

	provider := strings.ToLower(filteredArgs[1])
	competition := entity.Competition(0)

	switch strings.ToLower(filteredArgs[2]) {
	case "la_liga":
		competition = entity.LaLiga
	default:
		slog.Error("Unknown competition", "competition", args[2])
		return 1
	}

	if err := db.Init(); err != nil {
		slog.Error("Failed to initialize database", "error", err)
		return 1
	}
	defer db.Close()

	slog.Debug("Initializing connection to provider", "provider", strings.ToUpper(provider))

	switch provider {
	case "football_org":
		if err := football_org.Sync(competition); err != nil {
			slog.Error("Failed to sync Football Data API", "error", err)
			return 1
		}
	default:
		slog.Error("Unknown provider", "provider", provider)
		return 1
	}

	slog.Info("Operation completed successfully", "provider", provider)
	return 0
}
