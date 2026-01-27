package services

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
)

// LoadPrivateKey builds an ECDSA private key from a hex string (optional 0x prefix).
func LoadPrivateKey(hexKey string) (*ecdsa.PrivateKey, error) {
	hexKey = strings.TrimPrefix(hexKey, "0x")
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
