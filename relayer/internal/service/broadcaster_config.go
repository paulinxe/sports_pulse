package service

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"relayer/internal/config"
)

// MatchRegistry submitMatch + MatchAlreadySubmitted error ABI
const MATCH_REGISTRY_SUBMIT_MATCH_ABI = `[{"type":"function","name":"submitMatch","inputs":[{"name":"matchId","type":"bytes32","internalType":"bytes32"},{"name":"competitionId","type":"uint32","internalType":"uint32"},{"name":"homeTeamId","type":"uint32","internalType":"uint32"},{"name":"awayTeamId","type":"uint32","internalType":"uint32"},{"name":"homeTeamScore","type":"uint8","internalType":"uint8"},{"name":"awayTeamScore","type":"uint8","internalType":"uint8"},{"name":"matchDate","type":"uint32","internalType":"uint32"},{"name":"signature","type":"bytes","internalType":"bytes"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"error","name":"MatchAlreadySubmitted","inputs":[{"name":"matchId","type":"bytes32","internalType":"bytes32"}]}]`

// BroadcasterConfig holds pre-loaded RPC, contract, key, chain ID and ABI for broadcasting.
type BroadcasterConfig struct {
	RPCURL          string
	ContractAddress common.Address
	PrivateKey      *ecdsa.PrivateKey
	ChainID         *big.Int
	ContractABI     abi.ABI
}

func BuildBroadcasterConfig(envVars *config.EnvVars) (*BroadcasterConfig, error) {
	chainIDInt, err := strconv.ParseInt(envVars.ChainID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parse chain id: %w", err)
	}
	chainIDBigInt := big.NewInt(chainIDInt)

	contractABI, err := abi.JSON(strings.NewReader(MATCH_REGISTRY_SUBMIT_MATCH_ABI))
	if err != nil {
		return nil, fmt.Errorf("parse abi: %w", err)
	}

	return &BroadcasterConfig{
		RPCURL:          envVars.RPCURL,
		ContractAddress: envVars.ContractAddress,
		PrivateKey:      envVars.PrivateKey,
		ChainID:         chainIDBigInt,
		ContractABI:     contractABI,
	}, nil
}
