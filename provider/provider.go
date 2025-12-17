package main

import (
	"flag"
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
	// First argument is the program name (used in error messages), not related to positional args
	fs := flag.NewFlagSet("provider", flag.ContinueOnError)
	debugMode := fs.Bool("debug", false, "Enable debug logging")

	// Parse flags, which will consume --debug and similar flags
	if err := fs.Parse(args[1:]); err != nil {
		if err == flag.ErrHelp {
			return 0
		}

		slog.Error("Failed to parse flags", "error", err)
		return 1
	}

	positionalArgs := fs.Args()

	if len(positionalArgs) != 2 && len(positionalArgs) != 3 {
		slog.Error("Usage: <provider> <competition> [--debug]")
		return 1
	}

	// Set up debug logging if flag is present
	if *debugMode {
		slog.SetDefault(
			slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			})),
		)
	}

	competition := entity.Competition(0)
	switch strings.ToLower(positionalArgs[1]) {
	case "la_liga":
		competition = entity.LaLiga
	default:
		slog.Error("Unknown competition", "competition", positionalArgs[1])
		return 1
	}

	if err := db.Init(); err != nil {
		slog.Error("Failed to initialize database", "error", err)
		return 1
	}
	defer db.Close()

	provider := strings.ToLower(positionalArgs[0])
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
