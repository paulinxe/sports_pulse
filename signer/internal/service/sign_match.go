package service

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"os"
	"signer/internal/entity"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

const ORACLE_NAME = "SportsPulse"
const ORACLE_VERSION = "1"
const ORACLE_STRUCT_NAME = "Match"

var types *apitypes.Types
var domain *apitypes.TypedDataDomain

func SignMatch(match entity.Match, privKey *ecdsa.PrivateKey, chainId uint) (string, error) {
	matchId := match.CanonicalID
	if !strings.HasPrefix(matchId, "0x") {
		matchId = "0x" + matchId
	}

	message := apitypes.TypedData{
		Types:       *getTypes(),
		PrimaryType: ORACLE_STRUCT_NAME,
		Domain:      *getDomain(chainId),
		Message: map[string]any{
			"matchId":   matchId,
			"homeScore": big.NewInt(int64(match.HomeTeamScore)),
			"awayScore": big.NewInt(int64(match.AwayTeamScore)),
		},
	}

	// Compute struct hash (EIP-712)
	structHash, err := message.HashStruct(message.PrimaryType, message.Message)
	if err != nil {
		return "", fmt.Errorf("failed to hash struct: %w", err)
	}

	// Compute domain separator hash
	domainSeparator, err := message.HashStruct("EIP712Domain", message.Domain.Map())
	if err != nil {
		return "", fmt.Errorf("failed to hash domain: %w", err)
	}

	// Compute the final EIP-712 hash manually to match Solidity's MessageHashUtils.toTypedDataHash
	// Solidity does: keccak256(abi.encodePacked("\x19\x01", domainSeparator, structHash))
	finalHash := crypto.Keccak256Hash([]byte("\x19\x01"), domainSeparator, structHash)

	// Sign the hash
	signature, err := crypto.Sign(finalHash.Bytes(), privKey)
	if err != nil {
		return "", fmt.Errorf("sign hash: %w", err)
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

func getTypes() *apitypes.Types {
	if types != nil {
		return types
	}

	types = &apitypes.Types{
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

	return types
}

func getDomain(chainId uint) *apitypes.TypedDataDomain {
	if domain != nil {
		return domain
	}

	chainIdHexOrDecimal := math.NewHexOrDecimal256(int64(chainId))
	domain = &apitypes.TypedDataDomain{
		Name:              ORACLE_NAME,
		Version:           ORACLE_VERSION,
		ChainId:           chainIdHexOrDecimal,
		VerifyingContract: os.Getenv("ORACLE_CONTRACT_ADDRESS"),
	}

	return domain
}
