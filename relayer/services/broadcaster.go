package services

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log/slog"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"relayer/config"
	"relayer/entity"
)

// MatchRegistry submitMatch ABI (bytes32,uint32,uint32,uint32,uint8,uint8,uint32,bytes)
const matchRegistrySubmitMatchABI = `[{"type":"function","name":"submitMatch","inputs":[{"name":"matchId","type":"bytes32","internalType":"bytes32"},{"name":"competitionId","type":"uint32","internalType":"uint32"},{"name":"homeTeamId","type":"uint32","internalType":"uint32"},{"name":"awayTeamId","type":"uint32","internalType":"uint32"},{"name":"homeTeamScore","type":"uint8","internalType":"uint8"},{"name":"awayTeamScore","type":"uint8","internalType":"uint8"},{"name":"matchDate","type":"uint32","internalType":"uint32"},{"name":"signature","type":"bytes","internalType":"bytes"}],"outputs":[],"stateMutability":"nonpayable"}]`

type Broadcaster interface {
	Broadcast(ctx context.Context, match entity.Match) error
}

// BroadcastConfig holds pre-loaded RPC, contract, key, chain ID and ABI for broadcasting.
type BroadcasterConfig struct {
	RPCURL          string
	ContractAddress common.Address
	PrivateKey      *ecdsa.PrivateKey
	ChainID         *big.Int
	ContractABI     abi.ABI
}

func BuildBroadcasterConfig(envVars config.EnvVars) (BroadcasterConfig, error) {
	// TODO: see if we can simplify this conversion.
	chainIDInt, err := strconv.ParseInt(envVars.ChainID, 10, 64)
	if err != nil {
		return BroadcasterConfig{}, fmt.Errorf("parse chain id: %w", err)
	}
	chainIDBigInt := big.NewInt(chainIDInt)

	contractABI, err := abi.JSON(strings.NewReader(matchRegistrySubmitMatchABI))
	if err != nil {
		return BroadcasterConfig{}, fmt.Errorf("parse abi: %w", err)
	}

	return BroadcasterConfig{
		RPCURL:          envVars.RPCURL,
		ContractAddress: envVars.ContractAddress,
		PrivateKey:      envVars.PrivateKey,
		ChainID:         chainIDBigInt,
		ContractABI:     contractABI,
	}, nil
}

// BroadcastMatches runs Broadcast for each match in order (sequential for nonce safety).
// It returns the number of failed broadcasts.
func BroadcastMatches(broadcaster Broadcaster, matches []entity.Match, timeout time.Duration) (failedCount int) {
	for _, m := range matches {
		bctx, cancel := context.WithTimeout(context.Background(), timeout)
		err := broadcaster.Broadcast(bctx, m)
		cancel()

		if err != nil {
			// TODO: we need a new status for failed broadcasts so we can reconcile this later.
			slog.Error("broadcast failed", "match_id", m.ID, "canonical_id", m.CanonicalID, "error", err)
			failedCount++
			continue
		}

		slog.Info("broadcasted match", "canonical_id", m.CanonicalID)
	}

	return failedCount
}
