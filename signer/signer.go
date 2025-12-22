package main

import (
	"crypto/ecdsa"
	"encoding/asn1"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"log/slog"
	"math/big"
	"os"
	"signer/db"
	"signer/entity"
	"signer/repository"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

// ecPrivateKey represents the ASN.1 structure of an EC private key (SEC1 format)
type ecPrivateKey struct {
	Version       int
	PrivateKey    []byte
	NamedCurveOID asn1.ObjectIdentifier `asn1:"optional,explicit,tag:0"`
	PublicKey     asn1.BitString        `asn1:"optional,explicit,tag:1"`
}

// TODO: check if we can add more error codes.
// Check also in provider service.
type ErrorCodes int

const (
	_ ErrorCodes = iota
	DB_INIT_FAIL
	DB_QUERY_FAIL
	PRIVATE_KEY_LOAD_FAIL
	CHAIN_ID_NOT_VALID
)

const ORACLE_NAME = "SportsPulse"
const ORACLE_VERSION = "1"
const ORACLE_STRUCT_NAME = "Match"

func main() {
	os.Exit(Run())
}

func Run() int {
	// Check if database is already initialized (e.g., in tests)
	// TODO: check if this is really needed
	shouldClose := db.DB == nil
	if db.DB == nil {
		if err := db.Init(); err != nil {
			slog.Error("Failed to initialize database", "error", err)
			return int(DB_INIT_FAIL)
		}
	}
	if shouldClose {
		defer db.Close()
	}

	matches, err := repository.FindMatchesToSign()
	if err != nil {
		slog.Error("Failed to find matches to sign", "error", err)
		return int(DB_QUERY_FAIL)
	}

	slog.Debug(fmt.Sprintf("Found %d matches to sign", len(matches)))

	if len(matches) == 0 {
		return 0
	}

	privKey, err := loadPrivateKeyFromHex("0x4ba521e286bca3aa5fe1a8a8cf38017246b15fd2a4d9c79f1576ca82b9244279")
	slog.Info("privKey", "privKey", privKey)
	if err != nil {
		slog.Error("Failed to load private key", "error", err)
		return int(PRIVATE_KEY_LOAD_FAIL)
	}

	chainId, err := getChainId()
	if err != nil {
		slog.Error("Failed to get chain ID", "error", err)
		return int(CHAIN_ID_NOT_VALID)
	}

	// Declare EIP-712 type structure
	types := apitypes.Types{
		"EIP712Domain": []apitypes.Type{
			{Name: "name", Type: "string"},
			{Name: "version", Type: "string"},
			{Name: "chainId", Type: "uint256"},
			{Name: "verifyingContract", Type: "address"},
		},
		ORACLE_STRUCT_NAME: []apitypes.Type{
			{Name: "matchId", Type: "bytes32"},
			{Name: "homeScore", Type: "uint8"},
			{Name: "awayScore", Type: "uint8"},
		},
	}
	domain := apitypes.TypedDataDomain{
		Name:              ORACLE_NAME,
		Version:           ORACLE_VERSION,
		ChainId:           chainId,
		VerifyingContract: os.Getenv("ORACLE_CONTRACT_ADDRESS"),
	}

	for _, match := range matches {
		signature, err := signMatch(match, types, domain, privKey)
		if err != nil {
			slog.Error("Failed to sign match", "error", err, "match", match)
			continue
		}

		err = repository.StoreSignature(match, signature)
		if err != nil {
			slog.Error("Failed to store signature", "error", err, "match", match)
			continue
		}
	}

	return 0
}

// loadPrivateKeyFromHex creates an ECDSA private key from a hex-encoded string
// The hex string can have an optional 0x prefix
func loadPrivateKeyFromHex(hexKey string) (*ecdsa.PrivateKey, error) {
	// Remove 0x prefix if present
	hexKey = strings.TrimPrefix(hexKey, "0x")

	// Decode hex string to bytes
	privKeyBytes, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode hex private key: %v", err)
	}

	// Validate length (should be 32 bytes for secp256k1)
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

func loadPrivateKey(keyFile string) (*ecdsa.PrivateKey, error) {
	// TODO: simplify this. maybe generating keys with go-ethereum?
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

func getChainId() (*math.HexOrDecimal256, error) {
	chainIdStr := os.Getenv("CHAIN_ID")
	chainIdBig, ok := new(big.Int).SetString(chainIdStr, 10)
	if !ok {
		return nil, fmt.Errorf("invalid chain ID: %s", chainIdStr)
	}

	return (*math.HexOrDecimal256)(chainIdBig), nil
}

func signMatch(match entity.Match, types apitypes.Types, domain apitypes.TypedDataDomain, privKey *ecdsa.PrivateKey) (string, error) {
	homeScore := new(big.Int).SetUint64(uint64(match.HomeTeamScore))
	awayScore := new(big.Int).SetUint64(uint64(match.AwayTeamScore))

	// Convert matchId from hex string to bytes32
	var matchIdBytes [32]byte
	matchIdHex := strings.TrimPrefix(match.CanonicalID, "0x")
	matchIdBytesSlice, err := hex.DecodeString(matchIdHex)
	if err != nil {
		return "", fmt.Errorf("failed to decode matchId: %w", err)
	}
	if len(matchIdBytesSlice) != 32 {
		return "", fmt.Errorf("matchId must be 32 bytes, got %d", len(matchIdBytesSlice))
	}
	copy(matchIdBytes[:], matchIdBytesSlice)

	message := apitypes.TypedData{
		Types:       types,
		PrimaryType: ORACLE_STRUCT_NAME,
		Domain:      domain,
		Message: map[string]any{
			"matchId":   matchIdBytes, // Use bytes32 directly, not string
			"homeScore": homeScore,
			"awayScore": awayScore,
		},
	}

	// Log domain information for debugging
	domainMap := message.Domain.Map()
	slog.Info("=== Go EIP-712 Hash Computation ===")
	slog.Info("Domain map", "map", domainMap)
	slog.Info("Domain verifyingContract", "addr", domain.VerifyingContract)
	slog.Info("MatchId bytes32", "matchId", fmt.Sprintf("%x", matchIdBytes))
	slog.Info("MatchId (decimal)", "matchId", new(big.Int).SetBytes(matchIdBytes[:]).String())
	slog.Info("HomeScore", "score", match.HomeTeamScore)
	slog.Info("AwayScore", "score", match.AwayTeamScore)
	
	// Verify the type hash string matches Solidity
	// Solidity: keccak256("Match(bytes32 matchId,uint8 homeScore,uint8 awayScore)")
	expectedTypeHashStr := "Match(bytes32 matchId,uint8 homeScore,uint8 awayScore)"
	expectedTypeHash := crypto.Keccak256Hash([]byte(expectedTypeHashStr))
	slog.Info("Expected MATCH_RESULT_TYPEHASH", "hash", fmt.Sprintf("%x", expectedTypeHash))
	slog.Info("Expected MATCH_RESULT_TYPEHASH (decimal)", "hash", expectedTypeHash.Big().String())
	
	// Compute struct hash (EIP-712)
	// This should match: keccak256(abi.encode(MATCH_RESULT_TYPEHASH, matchId, homeScore, awayScore))
	structHash, err := message.HashStruct(message.PrimaryType, message.Message)
	if err != nil {
		return "", fmt.Errorf("failed to hash struct: %w", err)
	}
	slog.Info("Struct hash", "structHash", fmt.Sprintf("%x", structHash))
	slog.Info("Struct hash (decimal)", "structHash", new(big.Int).SetBytes(structHash).String())
	
	// Compute domain separator hash
	domainSeparator, err := message.HashStruct("EIP712Domain", domainMap)
	if err != nil {
		return "", fmt.Errorf("failed to hash domain: %w", err)
	}
	slog.Info("Domain separator", "separator", fmt.Sprintf("%x", domainSeparator))
	slog.Info("Domain separator (decimal)", "separator", new(big.Int).SetBytes(domainSeparator).String())
	
	// Compute the final EIP-712 hash manually to match Solidity's MessageHashUtils.toTypedDataHash
	// Solidity does: keccak256(abi.encodePacked("\x19\x01", domainSeparator, structHash))
	finalHash := crypto.Keccak256Hash(
		[]byte("\x19\x01"),
		domainSeparator,
		structHash,
	)
	slog.Info("Final hash", "hash", fmt.Sprintf("%x", finalHash))
	slog.Info("Final hash (decimal)", "hash", finalHash.Big().String())
	slog.Info("==================================")
	
	// Recover public key to verify
	pubKey := privKey.Public()
	pubKeyECDSA, ok := pubKey.(*ecdsa.PublicKey)
	if ok {
		expectedAddress := crypto.PubkeyToAddress(*pubKeyECDSA)
		slog.Info("Expected signer address", "address", expectedAddress.Hex())
	}

	// Sign the hash
	signature, err := crypto.Sign(finalHash.Bytes(), privKey)
	if err != nil {
		return "", err
	}

	// Adjust recovery ID: crypto.Sign returns 0 or 1, but Ethereum expects 27 or 28
	// For EIP-712, we use 27 or 28 (not EIP-155 adjusted, since chainId is in domain)
	recoveryID := signature[64]
	if recoveryID > 1 {
		return "", fmt.Errorf("invalid recovery ID from crypto.Sign: %d", recoveryID)
	}
	signature[64] = recoveryID + 27

	return common.Bytes2Hex(signature), nil
}
