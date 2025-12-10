package main

import (
    "log/slog"
    "os"
    "provider/db"
    "provider/football_org"
    "provider/entity"
    "strings"
)

func main() {
    os.Exit(run(os.Args))
}

func run(args []string) int {
    if len(args) != 3 {
        slog.Error("Usage: provider <provider> <competition>")
        return 1
    }

    provider := strings.ToLower(args[1])
    competition := entity.Competition(0)

    switch strings.ToLower(args[2]) {
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

    slog.Info("Initializing connection to provider", "provider", strings.ToUpper(provider))

    switch provider {
    case "football_org":
        slog.Info("Initializing Football Data API sync")
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
