package config

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type EnvVars struct {
	RPCURL          string
	ContractAddress common.Address
	PrivateKey      *ecdsa.PrivateKey
	ChainID         string
}

func LoadEnvVars() (*EnvVars, error) {
	rpcURL := os.Getenv("RPC_URL")
	if rpcURL == "" {
		return nil, fmt.Errorf("RPC_URL is not set")
	}

	contractAddr := os.Getenv("ORACLE_CONTRACT_ADDRESS")
	if contractAddr == "" {
		return nil, fmt.Errorf("ORACLE_CONTRACT_ADDRESS is not set")
	}
	if !common.IsHexAddress(contractAddr) {
		return nil, fmt.Errorf("invalid ORACLE_CONTRACT_ADDRESS: %s", contractAddr)
	}

	privateKey := os.Getenv("RELAYER_PRIVATE_KEY")
	if privateKey == "" {
		return nil, fmt.Errorf("RELAYER_PRIVATE_KEY is not set")
	}
	privateKeyParsed, err := parsePrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}

	chainID := os.Getenv("CHAIN_ID")
	if chainID == "" {
		return nil, fmt.Errorf("CHAIN_ID is not set")
	}

	return &EnvVars{
		RPCURL:          rpcURL,
		ContractAddress: common.HexToAddress(contractAddr),
		PrivateKey:      privateKeyParsed,
		ChainID:         chainID,
	}, nil
}

// parsePrivateKey builds an ECDSA private key from a hex string (optional 0x prefix).
func parsePrivateKey(privateKey string) (*ecdsa.PrivateKey, error) {
	hexKey := strings.TrimPrefix(privateKey, "0x")
	privKeyBytes, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, fmt.Errorf("decode hex private key: %w", err)
	}

	if len(privKeyBytes) != 32 {
		return nil, fmt.Errorf("invalid private key length: expected 32 bytes, got %d", len(privKeyBytes))
	}

	privKey, err := crypto.ToECDSA(privKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("create ECDSA key: %w", err)
	}

	if privKey.Curve != crypto.S256() {
		return nil, fmt.Errorf("private key is not secp256k1")
	}

	return privKey, nil
}
