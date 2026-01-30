package services

import (
	"context"
	"crypto/ecdsa"
	"errors"
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
	"relayer/repository"
)

// ErrMatchAlreadySubmitted is returned when the contract reverts with MatchAlreadySubmitted(bytes32).
var ErrMatchAlreadySubmitted = errors.New("match already submitted")

// MatchRegistry submitMatch + MatchAlreadySubmitted error ABI
const matchRegistrySubmitMatchABI = `[{"type":"function","name":"submitMatch","inputs":[{"name":"matchId","type":"bytes32","internalType":"bytes32"},{"name":"competitionId","type":"uint32","internalType":"uint32"},{"name":"homeTeamId","type":"uint32","internalType":"uint32"},{"name":"awayTeamId","type":"uint32","internalType":"uint32"},{"name":"homeTeamScore","type":"uint8","internalType":"uint8"},{"name":"awayTeamScore","type":"uint8","internalType":"uint8"},{"name":"matchDate","type":"uint32","internalType":"uint32"},{"name":"signature","type":"bytes","internalType":"bytes"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"error","name":"MatchAlreadySubmitted","inputs":[{"name":"matchId","type":"bytes32","internalType":"bytes32"}]}]`

type Broadcaster interface {
	Broadcast(ctx context.Context, calldata []byte) error
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
		calldata, err := buildCalldata(m)
		if err != nil {
			// This should never happen as the only reason for this to fail is if the ABI is invalid.
			// Anyways, lest's just log it and continue.
			slog.Error("build submitMatch calldata failed", "match_id", m.ID, "error", err)
			failedCount++
			cancel()
			continue
		}

		err = broadcaster.Broadcast(bctx, calldata)

		if err != nil {
			if errors.Is(err, ErrMatchAlreadySubmitted) {
				if updateErr := repository.BroadcastMatch(bctx, m.ID); updateErr != nil {
					slog.Error("broadcast match status update failed after already-submitted", "match_id", m.ID, "error", updateErr)
				}

				slog.Info("match already submitted on chain", "canonical_id", m.CanonicalID)
				cancel()
				continue
			}

			// TODO: we need a new status for failed broadcasts so we can reconcile this later.
			slog.Error("broadcast failed", "match_id", m.ID, "canonical_id", m.CanonicalID, "error", err)
			failedCount++
			cancel()
			continue
		}

		if updateErr := repository.BroadcastMatch(bctx, m.ID); updateErr != nil {
			// If we don't manage to update the status after a succesful broadcast, that is not an issue.
			// The next run will retry and contract will return MatchAlreadySubmitted which we are already handling above.
			slog.Error("broadcast match status update failed", "match_id", m.ID, "error", updateErr)
		} else {
			slog.Info("broadcasted match", "canonical_id", m.CanonicalID)
		}

		cancel()
	}

	return failedCount
}

func buildCalldata(match entity.Match) ([]byte, error) {
	contractABI, err := abi.JSON(strings.NewReader(matchRegistrySubmitMatchABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse abi: %w", err)
	}

	matchID := common.HexToHash(strings.TrimPrefix(match.CanonicalID, "0x"))
	return contractABI.Pack("submitMatch",
		matchID,
		uint32(match.CompetitionID),
		uint32(match.HomeTeamID),
		uint32(match.AwayTeamID),
		uint8(match.HomeTeamScore),
		uint8(match.AwayTeamScore),
		match.Start,
		match.Signature,
	)
}