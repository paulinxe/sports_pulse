package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"mock_api/db"
	"mock_api/repository"
)

func main() {
	os.Exit(run(os.Args))
}

func run(args []string) int {
	fs := flag.NewFlagSet("mock-api", flag.ExitOnError)
	serveFlag := fs.Bool("serve", false, "Start the mock API server")
	buildFlag := fs.Bool("build", false, "Initialize schedule and matches")

	if err := fs.Parse(args[1:]); err != nil {
		slog.Error("Failed to parse flags", "error", err)
		printUsage()
		return 1
	}

	if !*serveFlag && !*buildFlag {
		slog.Error("Must specify either --serve or --build")
		printUsage()
		return 1
	}

	if *serveFlag && *buildFlag {
		slog.Error("Cannot specify both --serve and --build")
		printUsage()
		return 1
	}

	if *serveFlag {
		return serve()
	}

	if *buildFlag {
		return build()
	}

	return 0
}

func serve() int {
	slog.Info("Starting mock API server")

	if err := db.Init(); err != nil {
		slog.Error("Failed to initialize database", "error", err)
		return 1
	}
	defer db.Close()

	if err := Start(); err != nil {
		slog.Error("Failed to start server", "error", err)
		return 1
	}

	return 0
}

func build() int {
	slog.Info("Building schedule and matches")

	if err := db.Init(); err != nil {
		slog.Error("Failed to initialize database", "error", err)
		return 1
	}
	defer db.Close()

	// Clear existing matches
	if err := repository.ClearAllMatches(); err != nil {
		slog.Error("Failed to clear matches", "error", err)
		return 1
	}

	// Generate schedule
	schedule, err := GenerateSchedule()
	if err != nil {
		slog.Error("Failed to generate schedule", "error", err)
		return 1
	}

	// Convert scheduled matches to matches with scores
	matches := make([]repository.Match, 0, len(schedule))
	for _, scheduled := range schedule {
		match := CreateMatch(scheduled)
		matches = append(matches, match)
	}

	// Insert matches into database
	if err := repository.InsertMatchesBatch(matches); err != nil {
		slog.Error("Failed to insert matches", "error", err)
		return 1
	}

	slog.Info("Schedule and matches initialized successfully", "matches_count", len(matches))
	return 0
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: mock-api [flags]\n\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	fmt.Fprintf(os.Stderr, "  --serve    Start the mock API server\n")
	fmt.Fprintf(os.Stderr, "  --build    Initialize schedule and matches\n")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "Examples:\n")
	fmt.Fprintf(os.Stderr, "  mock-api --serve\n")
	fmt.Fprintf(os.Stderr, "  mock-api --build\n")
}
