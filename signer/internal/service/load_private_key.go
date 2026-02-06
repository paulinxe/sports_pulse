package service

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
)

// Creates an ECDSA private key from a hex-encoded string
// The hex string can have an optional 0x prefix
func LoadPrivateKey(hexKey string) (*ecdsa.PrivateKey, error) {
	// Remove 0x prefix if present
	hexKey = strings.TrimPrefix(hexKey, "0x")

	// Decode hex string to bytes
	privKeyBytes, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode hex private key: %w", err)
	}

	// Validate length (should be 32 bytes for secp256k1)
	if len(privKeyBytes) != 32 {
		return nil, fmt.Errorf("invalid private key length: expected 32 bytes, got %d", len(privKeyBytes))
	}

	// Convert to ECDSA private key using go-ethereum's crypto
	privKey, err := crypto.ToECDSA(privKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to create ECDSA key: %w", err)
	}

	// Verify it's a secp256k1 key (required for Ethereum)
	if privKey.Curve != crypto.S256() {
		return nil, fmt.Errorf("private key is not secp256k1 (required for Ethereum)")
	}

	return privKey, nil
}
