package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	ethereum "github.com/ethereum/go-ethereum"
	"relayer/internal/entity"
	"relayer/internal/repository"
)

// GAS_ESTIMATE_FALLBACK_LIMIT is used when EstimateGas fails. Contract consumes ~120k; 150k provides a small buffer.
const GAS_ESTIMATE_FALLBACK_LIMIT = 130_000
// GAS_ESTIMATE_BUFFER_PERCENT is the percentage of the estimated gas to add as a buffer.
const GAS_ESTIMATE_BUFFER_PERCENT = 2

// ChainClient is the subset of chain operations needed for broadcasting and waiting for receipts.
type ChainClient interface {
	PendingNonceAt(ctx context.Context, account common.Address) (uint64, error)
	SuggestGasTipCap(ctx context.Context) (*big.Int, error)
	SuggestGasPrice(ctx context.Context) (*big.Int, error)
	EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error)
	SendTransaction(ctx context.Context, tx *types.Transaction) error
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error)
}

// ErrMatchAlreadySubmitted is returned when the contract reverts with MatchAlreadySubmitted(bytes32).
var ErrMatchAlreadySubmitted = errors.New("match already submitted")

// BroadcastMatches runs broadcast for each match in order (sequential for nonce safety).
// It returns the number of failed broadcasts.
func BroadcastMatches(client ChainClient, cfg BroadcasterConfig, repo *repository.MatchRepository, matches []entity.Match, timeout time.Duration) (failedCount int) {
	for _, match := range matches {
		bctx, cancel := context.WithTimeout(context.Background(), timeout)
		calldata, err := buildCalldata(match)
		if err != nil {
			slog.Error("build submitMatch calldata failed", "match_id", match.ID, "error", err)
			failedCount++
			cancel()
			continue
		}

		err = broadcast(bctx, client, cfg, calldata)

		if err != nil {
			if errors.Is(err, ErrMatchAlreadySubmitted) {
				if updateErr := repo.BroadcastMatch(bctx, match.ID); updateErr != nil {
					slog.Error("broadcast match status update failed after already-submitted", "match_id", match.ID, "error", updateErr)
				}

				slog.Info("match already submitted on chain", "canonical_id", match.CanonicalID)
				cancel()
				continue
			}

			slog.Error("broadcast failed", "match_id", match.ID, "canonical_id", match.CanonicalID, "error", err)
			failedCount++
			cancel()
			continue
		}

		if updateErr := repo.BroadcastMatch(bctx, match.ID); updateErr != nil {
			slog.Error("broadcast match status update failed", "match_id", match.ID, "error", updateErr)
		} else {
			slog.Info("broadcasted match", "canonical_id", match.CanonicalID)
		}

		cancel()
	}

	return failedCount
}

func buildCalldata(match entity.Match) ([]byte, error) {
	contractABI, err := abi.JSON(strings.NewReader(MATCH_REGISTRY_SUBMIT_MATCH_ABI))
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

func broadcast(ctx context.Context, client ChainClient, config BroadcasterConfig, calldata []byte) error {
	auth := crypto.PubkeyToAddress(config.PrivateKey.PublicKey)
	nonce, err := client.PendingNonceAt(ctx, auth)
	if err != nil {
		return fmt.Errorf("nonce: %w", err)
	}

	tipCap, err := client.SuggestGasTipCap(ctx)
	if err != nil {
		return fmt.Errorf("suggest gas tip: %w", err)
	}

	feeCap, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return fmt.Errorf("suggest gas price: %w", err)
	}

	estimatedGas, err := client.EstimateGas(ctx, ethereum.CallMsg{
		From:  auth,
		To:    &config.ContractAddress,
		Data:  calldata,
		Value: big.NewInt(0),
	})
	if err != nil {
		estimatedGas = uint64(GAS_ESTIMATE_FALLBACK_LIMIT)
		slog.Warn("gas estimation failed, using fallback limit", "error", err, "fallback_gas", GAS_ESTIMATE_FALLBACK_LIMIT)
	}

	gasLimit := estimatedGas + (estimatedGas * GAS_ESTIMATE_BUFFER_PERCENT / 100)

	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   config.ChainID,
		Nonce:     nonce,
		GasTipCap: tipCap,
		GasFeeCap: feeCap,
		Gas:       gasLimit,
		To:        &config.ContractAddress,
		Value:     big.NewInt(0),
		Data:      calldata,
	})
	signed, err := types.SignTx(tx, types.LatestSignerForChainID(config.ChainID), config.PrivateKey)
	if err != nil {
		return fmt.Errorf("sign tx: %w", err)
	}

	if err := client.SendTransaction(ctx, signed); err != nil {
		if isMatchAlreadySubmitted(err, config.ContractABI) {
			return ErrMatchAlreadySubmitted
		}

		return fmt.Errorf("send tx: %w", err)
	}

	receipt, err := waitForReceipt(ctx, client, signed.Hash())
	if err != nil {
		return fmt.Errorf("failed to wait for receipt: %w", err)
	}

	if receipt.Status != types.ReceiptStatusSuccessful {
		return fmt.Errorf("transaction reverted (status %d, tx %s)", receipt.Status, signed.Hash().Hex())
	}

	slog.Info("broadcasted match", "tx_hash", signed.Hash().Hex())
	return nil
}

type chainReceiptGetter interface {
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error)
}

func waitForReceipt(ctx context.Context, client chainReceiptGetter, txHash common.Hash) (*types.Receipt, error) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for {
		receipt, err := client.TransactionReceipt(ctx, txHash)
		if err == nil {
			return receipt, nil
		}

		if !errors.Is(err, ethereum.NotFound) {
			return nil, err
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
		}
	}
}
