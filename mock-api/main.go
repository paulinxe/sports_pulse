package main

import (
	"fmt"
	"log/slog"
	"os"

	"mock_api/db"
)

func main() {
	os.Exit(serve())
}

func serve() int {
	slog.Info("Starting mock API server")
	
	// Initialize database connection
	if err := db.Init(); err != nil {
		slog.Error("Failed to initialize database", "error", err)
		return 1
	}
	defer db.Close()
	
	// TODO: Initialize schedule and matches
	
	if err := Start(); err != nil {
		slog.Error("Failed to start server", "error", err)
		return 1
	}

	return 0
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: mock-api <command>\n\n")
	fmt.Fprintf(os.Stderr, "Commands:\n")
	fmt.Fprintf(os.Stderr, "  serve    Start the mock API server\n")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "Examples:\n")
	fmt.Fprintf(os.Stderr, "  mock-api serve\n")
}

