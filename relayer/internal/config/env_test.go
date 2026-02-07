package config

import (
	"encoding/hex"
	"os"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func Test_we_get_an_error_when_env_vars_are_not_valid(t *testing.T) {
	tests := []struct {
		name          string
		prepare       func(t *testing.T)
		expectedError string
	}{
		{"rpc url not set", func(t *testing.T) { _ = os.Unsetenv("RPC_URL") }, "RPC_URL is not set"},
		{"rpc url empty string", func(t *testing.T) { t.Setenv("RPC_URL", "") }, "RPC_URL is not set"},
		{"oracle contract address not set", func(t *testing.T) { _ = os.Unsetenv("ORACLE_CONTRACT_ADDRESS") }, "ORACLE_CONTRACT_ADDRESS is not set"},
		{"oracle contract address empty string", func(t *testing.T) { t.Setenv("ORACLE_CONTRACT_ADDRESS", "") }, "ORACLE_CONTRACT_ADDRESS is not set"},
		{"relayer private key not set", func(t *testing.T) { _ = os.Unsetenv("RELAYER_PRIVATE_KEY") }, "RELAYER_PRIVATE_KEY is not set"},
		{"relayer private key empty string", func(t *testing.T) { t.Setenv("RELAYER_PRIVATE_KEY", "") }, "RELAYER_PRIVATE_KEY is not set"},
		{"chain id not set", func(t *testing.T) { _ = os.Unsetenv("CHAIN_ID") }, "CHAIN_ID is not set"},
		{"chain id empty string", func(t *testing.T) { t.Setenv("CHAIN_ID", "") }, "CHAIN_ID is not set"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setAllValidEnv(t)
			tt.prepare(t)
			_, err := LoadEnvVars()
			if err == nil {
				t.Fatal("expected error")
			}

			if err.Error() != tt.expectedError {
				t.Errorf("expected error = %q, got %q", tt.expectedError, err.Error())
			}
		})
	}
}

func Test_we_get_an_error_when_oracle_contract_address_is_not_a_hex_address(t *testing.T) {
	setAllValidEnv(t)
	t.Setenv("ORACLE_CONTRACT_ADDRESS", "not-a-hex-address")
	_, err := LoadEnvVars()
	if err == nil {
		t.Fatal("expected error for invalid ORACLE_CONTRACT_ADDRESS")
	}

	if err.Error() != "invalid ORACLE_CONTRACT_ADDRESS: not-a-hex-address" {
		t.Errorf("expected error = %q, got %q", "invalid ORACLE_CONTRACT_ADDRESS: not-a-hex-address", err.Error())
	}
}

func Test_we_get_an_error_when_relayer_private_key_is_not_a_valid_hex_string(t *testing.T) {
	setAllValidEnv(t)
	t.Setenv("RELAYER_PRIVATE_KEY", "not-hex")
	_, err := LoadEnvVars()
	if err == nil {
		t.Fatal("expected error for invalid RELAYER_PRIVATE_KEY")
	}

	if !strings.Contains(err.Error(), "parse private key: decode hex private key") {
		t.Errorf("expected error to contain %q, got %q", "parse private key: decode hex private key", err.Error())
	}
}

func Test_we_get_an_error_when_relayer_private_key_is_not_32_bytes(t *testing.T) {
	setAllValidEnv(t)
	t.Setenv("RELAYER_PRIVATE_KEY", "0x0102")
	_, err := LoadEnvVars()
	if err == nil {
		t.Fatal("expected error for wrong length RELAYER_PRIVATE_KEY")
	}

	if err.Error() != "parse private key: invalid private key length: expected 32 bytes, got 2" {
		t.Errorf("expected error = %q, got %q", "parse private key: invalid private key length: expected 32 bytes, got 2", err.Error())
	}
}

func Test_we_get_all_env_vars_when_they_are_valid(t *testing.T) {
	setAllValidEnv(t)

	actual, err := LoadEnvVars()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if actual.RPCURL != "https://example.com" {
		t.Errorf("expected RPCURL = %q, got %q", "https://example.com", actual.RPCURL)
	}

	if actual.ContractAddress != common.HexToAddress("0x0000000000000000000000000000000000000001") {
		t.Errorf("expected ContractAddress = %s, got %s", "0x0000000000000000000000000000000000000001", actual.ContractAddress.Hex())
	}

	if actual.PrivateKey == nil {
		t.Errorf("expected PrivateKey to be non-nil, got nil")
	}

	if actual.ChainID != "31337" {
		t.Errorf("expected ChainID = %q, got %q", "31337", actual.ChainID)
	}
}

func setAllValidEnv(t *testing.T) {
	t.Helper()
	t.Setenv("RPC_URL", "https://example.com")
	t.Setenv("ORACLE_CONTRACT_ADDRESS", "0x0000000000000000000000000000000000000001")
	setValidPrivateKeyAndChainID(t)
	t.Setenv("CHAIN_ID", "31337")
}

func setValidPrivateKeyAndChainID(t *testing.T) {
	t.Helper()
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	hexKey := hex.EncodeToString(crypto.FromECDSA(key))
	t.Setenv("RELAYER_PRIVATE_KEY", "0x"+hexKey)
}
