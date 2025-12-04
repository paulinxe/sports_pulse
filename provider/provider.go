package main

import (
	"fmt"
	"log/slog"
	"os"
	"provider/football_org"
	"provider/db"
	"strings"
)

func main() {
	os.Exit(run(os.Args))
}

func run(args []string) int {
	if len(args) != 2 {
		slog.Error("Usage: provider <provider>")
		return 1
	}

	provider := strings.ToLower(args[1])

	if err := db.Init(); err != nil {
		slog.Error("Failed to initialize database", "error", err)
		return 1
	}
	defer db.Close()

	err := runProviderLogic(provider)
	if err != nil {
		slog.Error("Provider execution failed", "error", err)
		return 1
	}

	return 0
}

func runProviderLogic(provider string) error {
	slog.Info("Initializing connection to provider", "provider", strings.ToUpper(provider))

	switch provider {
	case "football_org":
		slog.Info("Initializing Football Data API sync")
		if err := football_org.Sync(); err != nil {
			return err
		}
	default:
		// Return an error if the provider is unknown
		return fmt.Errorf("unknown provider '%s'. Supported providers: aws, gcp, azure, localhost, football_data_org", provider)
	}

	slog.Info("Operation completed successfully", "provider", provider)
	return nil
}

func buildError(msg string) error {
	slog.Error("Error", "error", msg)
	return fmt.Errorf("%s", msg)
}
