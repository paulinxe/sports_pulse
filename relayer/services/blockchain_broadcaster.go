package services

import (
	"context"
	"encoding/hex"
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
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"

	ethereum "github.com/ethereum/go-ethereum"
)

// BlockchainBroadcaster sends signed matches to the MatchRegistry contract.
type BlockchainBroadcaster struct {
	config BroadcasterConfig
}

// NewBlockchainBroadcaster builds a broadcaster from config.
func NewBlockchainBroadcaster(cfg BroadcasterConfig) *BlockchainBroadcaster {
	return &BlockchainBroadcaster{config: cfg}
}

// Broadcast submits a single match to the chain via MatchRegistry.submitMatch using the pre-built calldata.
func (broadcaster *BlockchainBroadcaster) Broadcast(ctx context.Context, calldata []byte) error {
	// TODO: most probably we should open only one client and use it for all broadcasts.
	client, err := ethclient.DialContext(ctx, broadcaster.config.RPCURL)
	if err != nil {
		return fmt.Errorf("dial rpc: %w", err)
	}
	defer client.Close()

	auth := crypto.PubkeyToAddress(broadcaster.config.PrivateKey.PublicKey)
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

	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   broadcaster.config.ChainID,
		Nonce:     nonce,
		GasTipCap: tipCap,
		GasFeeCap: feeCap,
		Gas:       300_000, // TODO: we need to estimate the gas cost of the transaction.
		To:        &broadcaster.config.ContractAddress,
		Value:     big.NewInt(0),
		Data:      calldata,
	})

	signed, err := types.SignTx(tx, types.LatestSignerForChainID(broadcaster.config.ChainID), broadcaster.config.PrivateKey)
	if err != nil {
		return fmt.Errorf("sign tx: %w", err)
	}

	if err := client.SendTransaction(ctx, signed); err != nil {
		if isMatchAlreadySubmitted(err, broadcaster.config.ContractABI) {
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

// waitForReceipt polls for the transaction receipt until it is available or ctx is done.
// Only NotFound (receipt not yet available) is retried; any other error is returned immediately.
// The ticker is created once and sends every 500ms, so we only poll at that interval.
func waitForReceipt(ctx context.Context, client *ethclient.Client, txHash common.Hash) (*types.Receipt, error) {
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
			// Receipt not found yet; wait 500ms and loop again
		}
	}
}

func isMatchAlreadySubmitted(err error, contractABI abi.ABI) bool {
	var dataErr rpc.DataError
	if !errors.As(err, &dataErr) {
		return false
	}

	data := dataErr.ErrorData()
	if data == nil {
		return false
	}

	errorSelector, ok := data.(string)
	if !ok {
		return false
	}

	// We are just converting the error selector we got to bytes and the comparing with MatchAlreadySubmitted error selector.
	errorSelector = strings.TrimPrefix(errorSelector, "0x")
	errorSelectorBytes, decodeErr := hex.DecodeString(errorSelector)
	if decodeErr != nil || len(errorSelectorBytes) < 4 {
		return false
	}

	matchSubmittedError, ok := contractABI.Errors["MatchAlreadySubmitted"]
	if !ok {
		return false
	}

	return string(errorSelectorBytes[:4]) == string(matchSubmittedError.ID[:4])
}
