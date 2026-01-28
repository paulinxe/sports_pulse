package services

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"relayer/config"
)

func Test_we_get_an_error_when_chain_id_is_invalid(t *testing.T) {
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("generate key: %v", err)
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
		t.Fatalf("generate key: %v", err)
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
		t.Fatalf("generate key: %v", err)
	}

	envVars := config.EnvVars{
		RPCURL:          "https://rpc.example.com",
		ContractAddress: common.HexToAddress("0x0000000000000000000000000000000000000001"),
		PrivateKey:      key,
		ChainID:         "31337",
	}

	cfg, err := BuildBroadcasterConfig(envVars)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
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
