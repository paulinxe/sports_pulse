package main

import (
	"fmt"
	"os"
	"strings"
)

// ANSI color codes for pretty CLI output
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorCyan   = "\033[36m"
)

func main() {
	// 1. Direct argument parsing (No flags/help text for cron usage)
	// os.Args[0] is the program name, os.Args[1] is the first argument
	args := os.Args

	if len(args) < 2 {
		printError("Missing required argument: 'provider'")
		os.Exit(1)
	}

	// 2. Extract the provider argument
	// We convert to lowercase to make it case-insensitive
	provider := strings.ToLower(args[1])

	// 3. Run logic based on the provider
	err := runProviderLogic(provider)
	if err != nil {
		printError(err.Error())
		os.Exit(1)
	}
}

func runProviderLogic(provider string) error {
	fmt.Printf("%sInitializing connection to provider: %s...%s\n", ColorCyan, strings.ToUpper(provider), ColorReset)

	switch provider {
    case "football_org":
        fmt.Println("✅ Initializing Football Data API sync...")
        sync_football_org()
	default:
		// Return an error if the provider is unknown
		return fmt.Errorf("unknown provider '%s'. Supported providers: aws, gcp, azure, localhost, football_data_org", provider)
	}

	fmt.Printf("%s\nSuccess! Operation completed for %s.%s\n", ColorGreen, provider, ColorReset)
	return nil
}

func printError(msg string) {
	fmt.Fprintf(os.Stderr, "%sError: %s%s\n", ColorRed, msg, ColorReset)
}
