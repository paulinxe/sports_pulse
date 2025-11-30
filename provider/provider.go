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
	os.Exit(run(os.Args))
}

func run(args []string) int {
	if len(args) != 2 {
		printError("Usage: provider <provider>")
		return 1
	}

	provider := strings.ToLower(args[1])

	err := runProviderLogic(provider)
	if err != nil {
		printError(err.Error())
		return 1
	}

	return 0
}

func runProviderLogic(provider string) error {
	fmt.Printf("%sInitializing connection to provider: %s...%s\n", ColorCyan, strings.ToUpper(provider), ColorReset)

	switch provider {
	case "football_org":
		fmt.Println("✅ Initializing Football Data API sync...")
		if err := sync_football_org(); err != nil {
			return err
		}
	default:
		// Return an error if the provider is unknown
		return fmt.Errorf("unknown provider '%s'. Supported providers: aws, gcp, azure, localhost, football_data_org", provider)
	}

	fmt.Printf("%s\nSuccess! Operation completed for %s.%s\n", ColorGreen, provider, ColorReset)
	return nil
}

func buildError(msg string) error {
	return fmt.Errorf("%sError: %s%s", ColorRed, msg, ColorReset)
}

func printError(msg string) {
	fmt.Fprintf(os.Stderr, "%sError: %s%s\n", ColorRed, msg, ColorReset)
}
