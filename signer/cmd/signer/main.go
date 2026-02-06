package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"signer/internal/config"
	"signer/internal/repository"
	"signer/internal/service"
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
	db, err := config.InitDB()
	if err != nil {
		slog.Error("failed to initialize database", "error", err)
		os.Exit(int(DB_INIT_FAIL))
	}
	defer func() { _ = db.Close() }()
	os.Exit(Run(db, DB_TIMEOUT, STORE_TIMEOUT))
}

func Run(db *sql.DB, dbTimeout, storeTimeout time.Duration) int {
	dbContext, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	repo, err := repository.NewMatchRepository(db)
	if err != nil {
		slog.Error("failed to create repository", "error", err)
		return int(DB_INIT_FAIL)
	}
	matches, err := repo.FindMatchesToSign(dbContext)
	if err != nil {
		slog.Error("failed to find matches to sign", "error", err)
		return int(DB_QUERY_FAIL)
	}

	slog.Debug(fmt.Sprintf("Found %d matches to sign", len(matches)))

	if len(matches) == 0 {
		return int(SUCCESS)
	}

	privateKey, err := service.LoadPrivateKey(os.Getenv("SIGNER_PRIVATE_KEY"))
	if err != nil {
		slog.Error("failed to load private key", "error", err)
		return int(PRIVATE_KEY_LOAD_FAIL)
	}

	chainId, err := service.LoadChainId()
	if err != nil {
		slog.Error("failed to get chain ID", "error", err)
		return int(CHAIN_ID_NOT_VALID)
	}

	for _, match := range matches {
		signature, err := service.SignMatch(match, privateKey, chainId)
		if err != nil {
			slog.Error("failed to sign match", "error", err, "match", match)
			continue
		}

		storeContext, storeCancel := context.WithTimeout(context.Background(), storeTimeout)
		err = repo.StoreSignature(storeContext, match, signature)
		storeCancel()

		if err != nil {
			slog.Error("failed to store signature", "error", err, "match", match)
			continue
		}
	}

	return int(SUCCESS)
}
