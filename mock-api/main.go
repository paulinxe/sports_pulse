package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"mock_api/db"
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
	
	schedule, err := GenerateSchedule()
	if err != nil {
		slog.Error("Failed to generate schedule", "error", err)
		return 1
	}

	for _, match := range schedule {
		fmt.Printf("Date: %s\n", match.Date.Format("2006-01-02"))
		fmt.Printf("Home Team ID: %d\n", match.HomeTeamID)
		fmt.Printf("Away Team ID: %d\n", match.AwayTeamID)
	}
	
	slog.Info("Schedule and matches initialized successfully")
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

