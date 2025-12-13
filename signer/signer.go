package main

import (
	"crypto/ecdsa"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	//"math/big"
	"os"
	"log/slog"
	"signer/db"
	"signer/repository"

	//"github.com/ethereum/go-ethereum/common"
	//"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	//"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

// ecPrivateKey represents the ASN.1 structure of an EC private key (SEC1 format)
type ecPrivateKey struct {
	Version       int
	PrivateKey    []byte
	NamedCurveOID asn1.ObjectIdentifier `asn1:"optional,explicit,tag:0"`
	PublicKey     asn1.BitString         `asn1:"optional,explicit,tag:1"`
}

func main() {
	os.Exit(Run())
}

func Run() int {
	if err := db.Init(); err != nil {
		slog.Error("Failed to initialize database", "error", err)
		return 1
	}
	defer db.Close()

	matches, err := repository.FindMatchesToSign()
	if err != nil {
		slog.Error("Failed to find matches to sign", "error", err)
		return 1
	}

	slog.Debug(fmt.Sprintf("Found %d matches to sign", len(matches)))

	if len(matches) == 0 {
		return 0
	}

	privKey, err := loadPrivateKey(os.Getenv("PRIVATE_KEY_FILE"))
	if err != nil {
		slog.Error("Failed to load private key", "error", err)
		return 1
	}

	for _, match := range matches {
		fmt.Println(match)
		fmt.Println(privKey)
	}

	return 0
}

func loadPrivateKey(keyFile string) (*ecdsa.PrivateKey, error) {
	keyBytes, err := os.ReadFile(keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %v", err)
	}

	// Parse PEM format (SEC1: -----BEGIN EC PRIVATE KEY-----)
	block, _ := pem.Decode(keyBytes)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	if block.Type != "EC PRIVATE KEY" {
		return nil, fmt.Errorf("expected EC PRIVATE KEY, got %s", block.Type)
	}

	// Parse SEC1 DER-encoded structure
	var ecPrivKey ecPrivateKey
	_, err = asn1.Unmarshal(block.Bytes, &ecPrivKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SEC1 structure: %v", err)
	}

	// Extract the raw private key bytes (should be 32 bytes for secp256k1)
	privKeyBytes := ecPrivKey.PrivateKey
	if len(privKeyBytes) != 32 {
		return nil, fmt.Errorf("invalid private key length: expected 32 bytes, got %d", len(privKeyBytes))
	}

	// Convert to ECDSA private key using go-ethereum's crypto
	privKey, err := crypto.ToECDSA(privKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to create ECDSA key: %v", err)
	}

	// Verify it's a secp256k1 key (required for Ethereum)
	if privKey.Curve != crypto.S256() {
		return nil, fmt.Errorf("private key is not secp256k1 (required for Ethereum)")
	}

	return privKey, nil
}

// func main() {
// 	if err := db.Init(); err != nil {
// 		slog.Error("Failed to initialize database", "error", err)
// 		os.Exit(1)
// 	}
// 	defer db.Close()

// 	matches, err := repository.FindMatchesToSign()
// 	if err != nil {
// 		slog.Error("Failed to find matches to sign", "error", err)
// 		os.Exit(1)
// 	}

	
// 	privKey, err := loadPrivateKey(os.Getenv("PRIVATE_KEY_FILE"))
// 	if err != nil {
// 		panic(err)
// 	}

// 	// EIP712 domain
// 	chainIdStr := os.Getenv("CHAIN_ID")
// 	chainIdBig, ok := new(big.Int).SetString(chainIdStr, 10)
// 	if !ok {
// 		panic("invalid chain ID: " + chainIdStr)
// 	}
// 	chainId := (*math.HexOrDecimal256)(chainIdBig)
// 	domain := apitypes.TypedDataDomain{
// 		Name:              "FanTokenPulse",
// 		Version:           "1",
// 		ChainId:           chainId,
// 		VerifyingContract: os.Getenv("ORACLE_CONTRACT_ADDRESS"), // Replace
// 	}

// 	// Declare EIP-712 type structure
// 	types := apitypes.Types{
// 		"MatchResult": []apitypes.Type{
// 			{Name: "matchId", Type: "uint256"},
// 			{Name: "homeScore", Type: "uint8"},
// 			{Name: "awayScore", Type: "uint8"},
// 			{Name: "timestamp", Type: "uint256"},
// 		},
// 	}

// 	// The actual message to sign
// 	message := apitypes.TypedData{
// 		Types:       types,
// 		PrimaryType: "MatchResult",
// 		Domain:      domain,
// 		Message: map[string]any{
// 			"matchId":   big.NewInt(123456),
// 			"homeScore": big.NewInt(3),
// 			"awayScore": big.NewInt(1),
// 			"timestamp": big.NewInt(1737492000),
// 		},
// 	}

// 	// Compute digest (EIP-712)
// 	digest, err := message.HashStruct(message.PrimaryType, message.Message)
// 	if err != nil {
// 		panic(err)
// 	}
// 	domainSeparator, _ := message.HashStruct("EIP712Domain", message.Domain.Map())
// 	finalHash := crypto.Keccak256Hash(
// 		[]byte("\x19\x01"),
// 		domainSeparator,
// 		digest,
// 	)

// 	// Sign the hash
// 	signature, err := crypto.Sign(finalHash.Bytes(), privKey)
// 	if err != nil {
// 		panic(err)
// 	}

// 	fmt.Println("Signature:", common.Bytes2Hex(signature))
// }
