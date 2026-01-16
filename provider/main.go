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
	UNKNOWN_OPERATION
	UNKNOWN_PROVIDER
	UNKNOWN_COMPETITION
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

	if len(args) < 3 {
		slog.Error("Usage: <operation> <provider> [competition]")
		slog.Error("Operations: sync, reconcile")
		slog.Error("For sync: sync <provider> <competition>")
		slog.Error("For reconcile: reconcile <provider>")
		return int(BAD_ARGUMENTS)
	}

	if err := db.Init(); err != nil {
		slog.Error("Failed to initialize database", "error", err)
		return int(DB_INIT_ERROR)
	}
	defer db.Close()

	operation := strings.ToLower(args[1])
	provider := args[2]
	slog.Debug("Initializing operation", "operation", strings.ToUpper(operation), "provider", strings.ToUpper(provider))

	var err error
	switch operation {
	case "sync":
		if len(args) != 4 {
			slog.Error("Usage for sync: sync <provider> <competition>")
			return int(BAD_ARGUMENTS)
		}

		err = Sync(provider, args[3])

	case "reconcile":
		if len(args) != 3 {
			slog.Error("Usage for reconcile: reconcile <provider>")
			return int(BAD_ARGUMENTS)
		}

		err = Reconcile(provider)

	default:
		slog.Error("Unknown operation", "operation", operation)
		slog.Error("Valid operations: sync, reconcile")
		return int(UNKNOWN_OPERATION)
	}

	if err != nil {
		slog.Error("Failed to complete operation", "operation", operation, "provider", provider, "error", err)
		return int(PROVIDER_ERROR)
	}

	slog.Info("Operation completed successfully", "operation", operation, "provider", provider)
	return int(SUCCESS)
}
