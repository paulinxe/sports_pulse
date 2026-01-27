package services

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"relayer/entity"
	"relayer/config"
)

// MatchRegistry submitMatch ABI (bytes32,uint32,uint32,uint32,uint8,uint8,uint32,bytes)
const matchRegistrySubmitMatchABI = `[{"inputs":[{"internalType":"bytes32","name":"matchId","type":"bytes32"},{"internalType":"uint32","name":"competitionId","type":"uint32"},{"internalType":"uint32","name":"homeTeamId","type":"uint32"},{"internalType":"uint32","name":"awayTeamId","type":"uint32"},{"internalType":"uint8","name":"homeTeamScore","type":"uint8"},{"internalType":"uint8","name":"awayTeamScore","type":"uint8"},{"internalType":"uint32","name":"matchDate","type":"uint32"},{"internalType":"bytes","name":"signature","type":"bytes"}],"name":"submitMatch","outputs":[],"stateMutability":"nonpayable","type":"function"}]`

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

	// TODO: should we return BroadcasterConfig or &BroadcasterConfig?
	return BroadcasterConfig{
		RPCURL:          envVars.RPCURL,
		ContractAddress: envVars.ContractAddress,
		PrivateKey:      envVars.PrivateKey,
		ChainID:         chainIDBigInt,
		ContractABI:     contractABI,
	}, nil
}
