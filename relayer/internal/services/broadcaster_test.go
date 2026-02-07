package services

import (
	"context"
	"encoding/hex"
	"errors"
	"math/big"
	"strings"
	"testing"
	"time"

	"relayer/internal/config"
	"relayer/internal/entity"
	"relayer/internal/repository"
	"relayer/testutil"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
)

func Test_we_get_an_error_when_chain_id_is_invalid(t *testing.T) {
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}

	envVars := config.EnvVars{
		RPCURL:          "https://rpc.example.com",
		ContractAddress: common.HexToAddress("0x0000000000000000000000000000000000000001"),
		PrivateKey:      key,
		ChainID:         "not-a-number",
	}

	_, err = BuildBroadcasterConfig(envVars)
	if err == nil {
		t.Fatal("expected error for invalid chain id")
	}

	if err.Error() != "parse chain id: strconv.ParseInt: parsing \"not-a-number\": invalid syntax" {
		t.Errorf("expected parse chain id error, got %q", err.Error())
	}
}

func Test_we_get_an_error_when_chain_id_is_empty(t *testing.T) {
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}

	envVars := config.EnvVars{
		RPCURL:          "https://rpc.example.com",
		ContractAddress: common.HexToAddress("0x0000000000000000000000000000000000000001"),
		PrivateKey:      key,
		ChainID:         "",
	}

	_, err = BuildBroadcasterConfig(envVars)
	if err == nil {
		t.Fatal("expected error for empty chain id")
	}

	if err.Error() != "parse chain id: strconv.ParseInt: parsing \"\": invalid syntax" {
		t.Errorf("expected parse chain id error for empty string, got %q", err.Error())
	}
}

func Test_we_get_a_broadcaster_config_when_all_env_vars_are_valid(t *testing.T) {
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}

	envVars := config.EnvVars{
		RPCURL:          "https://rpc.example.com",
		ContractAddress: common.HexToAddress("0x0000000000000000000000000000000000000001"),
		PrivateKey:      key,
		ChainID:         "31337",
	}

	cfg, err := BuildBroadcasterConfig(envVars)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.RPCURL != envVars.RPCURL {
		t.Errorf("RPCURL: expected %q, got %q", envVars.RPCURL, cfg.RPCURL)
	}

	if cfg.ContractAddress != envVars.ContractAddress {
		t.Errorf("ContractAddress: expected %s, got %s", envVars.ContractAddress.Hex(), cfg.ContractAddress.Hex())
	}

	if cfg.PrivateKey != envVars.PrivateKey {
		t.Error("PrivateKey: expected same pointer as envVars.PrivateKey")
	}

	if cfg.ChainID == nil || cfg.ChainID.Cmp(big.NewInt(31337)) != 0 {
		t.Errorf("ChainID: expected 31337, got %v", cfg.ChainID)
	}

	if cfg.ContractABI.Methods["submitMatch"].Name != "submitMatch" {
		t.Error("ContractABI: expected submitMatch method to be parsed")
	}
}

func Test_broadcast_uses_gas_estimate_from_client(t *testing.T) {
	cfg, mockClient, matches, repo := setupTest(t)
	mockedGasEstimate := uint64(100_000)
	mockClient.EstimatedGas = mockedGasEstimate
	mockClient.Receipt = &types.Receipt{Status: types.ReceiptStatusSuccessful}

	BroadcastMatches(mockClient, cfg, repo, matches, 30*time.Second)

	if mockClient.LastSentTx == nil {
		t.Fatal("expected transaction to be sent")
	}

	// 100_000 + 2% buffer
	expectedGas := uint64(102_000)
	if mockClient.LastSentTx.Gas() != expectedGas {
		t.Errorf("expected gas %d (estimate + buffer), got %d", expectedGas, mockClient.LastSentTx.Gas())
	}
}

func Test_broadcast_uses_fallback_gas_when_estimation_fails(t *testing.T) {
	cfg, mockClient, matches, repo := setupTest(t)
	mockClient.EstimateGasErr = errors.New("execution reverted")
	mockClient.Receipt = &types.Receipt{Status: types.ReceiptStatusSuccessful}

	BroadcastMatches(mockClient, cfg, repo, matches, 30*time.Second)

	if mockClient.LastSentTx == nil {
		t.Fatal("expected transaction to be sent")
	}

	// Fallback limit + buffer (130_000 + 2% = 132_600)
	expectedGas := uint64(132_600)
	if mockClient.LastSentTx.Gas() != expectedGas {
		t.Errorf("expected gas %d (fallback + buffer), got %d", expectedGas, mockClient.LastSentTx.Gas())
	}
}

func setupTest(t *testing.T) (BroadcasterConfig, *testutil.MockChainClient, []entity.Match, *repository.MatchRepository) {
	t.Helper()
	db, repo := testutil.InitDB(t)
	t.Cleanup(func() { testutil.CloseDB(db) })

	u := uuid.New()
	canonicalID := "0x" + hex.EncodeToString(u[:])
	start := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	testutil.InsertSignedMatch(t, db, uuid.New(), canonicalID, 1, 10, 20, 2, 1, start, "deadbeef")

	cfg, mockClient := buildBroadcastTestConfig(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	matches, err := repo.FindSignedMatches(ctx)
	cancel()

	if err != nil {
		t.Fatalf("find signed matches: %v", err)
	}

	return cfg, mockClient, matches, repo
}

func buildBroadcastTestConfig(t *testing.T) (BroadcasterConfig, *testutil.MockChainClient) {
	t.Helper()
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}

	contractABI, err := abi.JSON(strings.NewReader(MATCH_REGISTRY_SUBMIT_MATCH_ABI))
	if err != nil {
		t.Fatal(err)
	}

	return BroadcasterConfig{
			RPCURL:          "",
			ContractAddress: common.HexToAddress("0x0000000000000000000000000000000000000001"),
			PrivateKey:      key,
			ChainID:         big.NewInt(31337),
			ContractABI:     contractABI,
		}, &testutil.MockChainClient{
			Nonce:  0,
			TipCap: big.NewInt(1),
			FeeCap: big.NewInt(2),
		}
}
