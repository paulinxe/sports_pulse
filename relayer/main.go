package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"relayer/config"
	"relayer/repository"
	"relayer/services"
)

const (
	dbTimeout    = 30 * time.Second
	broadcastCtx = 60 * time.Second
)

type ErrorCodes int

const (
	SUCCESS ErrorCodes = iota
	DB_INIT_ERROR
	MISSING_ENV
	BUILD_CONFIG_ERROR
	FIND_MATCHES_ERROR
	BROADCAST_FAILURE
)

func main() {
	envVars, err := config.LoadEnvVars()
	if err != nil {
		slog.Error("failed to load environment variables", "error", err)
		os.Exit(int(MISSING_ENV))
	}

	cfg, err := services.BuildBroadcasterConfig(envVars)
	if err != nil {
		slog.Error("failed to build broadcast config", "error", err)
		os.Exit(int(BUILD_CONFIG_ERROR))
	}

	client, err := ethclient.DialContext(context.Background(), cfg.RPCURL)
	if err != nil {
		slog.Error("failed to dial chain", "error", err)
		os.Exit(int(BUILD_CONFIG_ERROR))
	}
	defer client.Close()

	os.Exit(Run(client, cfg))
}

func Run(client services.ChainClient, cfg services.BroadcasterConfig) int {
	shouldClose, err := config.InitDB()
	if err != nil {
		slog.Error("failed to initialize database", "error", err)
		return int(DB_INIT_ERROR)
	}

	if shouldClose {
		defer func() { _ = config.Close() }()
	}

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	matches, err := repository.FindSignedMatches(ctx)
	cancel()
	if err != nil {
		slog.Error("failed to find signed matches", "error", err)
		return int(FIND_MATCHES_ERROR)
	}

	slog.Info("found signed matches to broadcast", "count", len(matches))

	failed := services.BroadcastMatches(client, cfg, matches, broadcastCtx)
	if failed > 0 {
		return int(BROADCAST_FAILURE)
	}

	return int(SUCCESS)
}
