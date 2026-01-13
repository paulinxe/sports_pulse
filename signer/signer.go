package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"signer/db"
	"signer/repository"
	"signer/services"
	"time"
)

type ErrorCodes int

const (
	SUCCESS ErrorCodes = iota
	DB_INIT_FAIL
	DB_QUERY_FAIL
	PRIVATE_KEY_LOAD_FAIL
	CHAIN_ID_NOT_VALID
)

const (
	DB_TIMEOUT    = 30 * time.Second
	STORE_TIMEOUT = 10 * time.Second
)

func main() {
	os.Exit(Run(DB_TIMEOUT, STORE_TIMEOUT))
}

func Run(dbTimeout, storeTimeout time.Duration) int {
	shouldClose, err := db.Init()
	if err != nil {
		slog.Error("Failed to initialize database", "error", err)
		return int(DB_INIT_FAIL)
	}

	if shouldClose {
		defer db.Close()
	}

	dbContext, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	matches, err := repository.FindMatchesToSign(dbContext)
	if err != nil {
		slog.Error("Failed to find matches to sign", "error", err)
		return int(DB_QUERY_FAIL)
	}

	slog.Debug(fmt.Sprintf("Found %d matches to sign", len(matches)))

	if len(matches) == 0 {
		return int(SUCCESS)
	}

	privateKey, err := services.LoadPrivateKey(os.Getenv("SIGNER_PRIVATE_KEY"))
	if err != nil {
		slog.Error("Failed to load private key", "error", err)
		return int(PRIVATE_KEY_LOAD_FAIL)
	}

	chainId, err := services.LoadChainId()
	if err != nil {
		slog.Error("Failed to get chain ID", "error", err)
		return int(CHAIN_ID_NOT_VALID)
	}

	for _, match := range matches {
		signature, err := services.SignMatch(match, privateKey, chainId)
		if err != nil {
			slog.Error("Failed to sign match", "error", err, "match", match)
			continue
		}

		storeContext, storeCancel := context.WithTimeout(context.Background(), storeTimeout)
		defer storeCancel()

		err = repository.StoreSignature(storeContext, match, signature)

		if err != nil {
			slog.Error("Failed to store signature", "error", err, "match", match)
			continue
		}
	}

	return int(SUCCESS)
}
