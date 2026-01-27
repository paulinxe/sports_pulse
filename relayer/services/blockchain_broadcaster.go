package services

import (
	"context"
	"fmt"
	"log/slog"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"relayer/entity"
)

// BlockchainBroadcaster sends signed matches to the MatchRegistry contract.
type BlockchainBroadcaster struct {
	config BroadcasterConfig
}

// NewBlockchainBroadcaster builds a broadcaster from config.
func NewBlockchainBroadcaster(cfg BroadcasterConfig) *BlockchainBroadcaster {
	return &BlockchainBroadcaster{config: cfg}
}

// Broadcast submits a single match to the chain via MatchRegistry.submitMatch.
func (broadcaster *BlockchainBroadcaster) Broadcast(ctx context.Context, match entity.Match) error {
	// TODO: most probably we should open only one client and use it for all broadcasts.
	client, err := ethclient.DialContext(ctx, broadcaster.config.RPCURL)
	if err != nil {
		return fmt.Errorf("dial rpc: %w", err)
	}
	defer client.Close()

	matchID := common.HexToHash(strings.TrimPrefix(match.CanonicalID, "0x"))
	calldata, err := broadcaster.config.ContractABI.Pack("submitMatch",
		matchID,
		uint32(match.CompetitionID),
		uint32(match.HomeTeamID),
		uint32(match.AwayTeamID),
		uint8(match.HomeTeamScore),
		uint8(match.AwayTeamScore),
		match.Start,
		match.Signature,
	)
	if err != nil {
		return fmt.Errorf("pack submitMatch: %w", err)
	}

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
		return fmt.Errorf("send tx: %w", err)
	}

	slog.Info("broadcasted match", "canonical_id", match.CanonicalID, "tx_hash", signed.Hash().Hex())
	return nil
}
