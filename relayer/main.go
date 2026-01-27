package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"relayer/db"
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
	INVALID_CONTRACT_ADDRESS
	BUILD_CONFIG_ERROR
	FIND_MATCHES_ERROR
)

func main() {
	os.Exit(Run())
}

func Run() int {
	shouldClose, err := db.Init()
	if err != nil {
		slog.Error("failed to initialize database", "error", err)
		return int(DB_INIT_ERROR)
	}

	if shouldClose {
		defer func() { _ = db.Close() }()
	}

	rpcURL := os.Getenv("RPC_URL")
	if rpcURL == "" {
		slog.Error("missing env: RPC_URL, ORACLE_CONTRACT_ADDRESS, or RELAYER_PRIVATE_KEY")
		return int(MISSING_ENV)
	}

	contractAddr := os.Getenv("ORACLE_CONTRACT_ADDRESS")
	if contractAddr == "" {
		slog.Error("missing env: ORACLE_CONTRACT_ADDRESS")
		return int(MISSING_ENV)
	}

	relayerKey := os.Getenv("RELAYER_PRIVATE_KEY")
	if relayerKey == "" {
		slog.Error("missing env: RELAYER_PRIVATE_KEY")
		return int(MISSING_ENV)
	}

	chainID := os.Getenv("CHAIN_ID")
	if chainID == "" {
		slog.Error("missing env: CHAIN_ID")
		return int(MISSING_ENV)
	}

	if !common.IsHexAddress(contractAddr) {
		slog.Error("invalid ORACLE_CONTRACT_ADDRESS", "value", contractAddr)
		return int(INVALID_CONTRACT_ADDRESS)
	}

	cfgCtx, cfgCancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cfgCancel()
	cfg, err := services.BuildBroadcasterConfig(cfgCtx, rpcURL, chainID, common.HexToAddress(contractAddr), relayerKey)
	if err != nil {
		slog.Error("failed to build broadcast config", "error", err)
		return int(BUILD_CONFIG_ERROR)
	}

	broadcaster := services.NewBlockchainBroadcaster(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()
	matches, err := repository.FindSignedMatches(ctx)
	if err != nil {
		slog.Error("failed to find signed matches", "error", err)
		return int(FIND_MATCHES_ERROR)
	}

	slog.Info("found signed matches to broadcast", "count", len(matches))

	// TODO: we can't spawn too many goroutines, we need to limit the number of concurrent broadcasts.
	for _, m := range matches {
		func() {
			bctx, bcancel := context.WithTimeout(context.Background(), broadcastCtx)
			defer bcancel()
			err := broadcaster.Broadcast(bctx, m)
			if err != nil {
				slog.Error("broadcast failed", "match_id", m.ID, "canonical_id", m.CanonicalID, "error", err)
				return
			}
			slog.Info("broadcasted match", "canonical_id", m.CanonicalID)
		}()
	}

	return int(SUCCESS)
}
