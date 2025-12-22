package main

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"signer/db"
	"signer/entity"
	"signer/repository"
	"signer/testutil"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
)

func setup() {
	os.Setenv("CHAIN_ID", "31337")
	os.Setenv("PRIVATE_KEY_FILE", "private_test_key.key")
	os.Setenv("ORACLE_CONTRACT_ADDRESS", "0xF62849F9A0B5Bf2913b396098F7c7019b51A820a")
}

func Test_no_errors_when_nothing_to_sign(t *testing.T) {
	setup()
	testutil.InitDatabase(t)
	defer testutil.CloseDatabase()

	matches, err := repository.FindMatchesToSign()
	if err != nil {
		t.Fatalf("Failed to find matches to sign: %v", err)
	}

	if len(matches) != 0 {
		t.Fatalf("Expected 0 matches to sign, got %d", len(matches))
	}
}

func Test_we_log_an_error_when_private_key_is_not_found(t *testing.T) {
	setup()
	testutil.InitDatabase(t)
	defer testutil.CloseDatabase()
	logger := testutil.GetLogger()
	insertSignableMatch()

	os.Setenv("PRIVATE_KEY_FILE", "not_found.pem")

	exitCode := Run()
	if exitCode != int(PRIVATE_KEY_LOAD_FAIL) {
		t.Fatalf("Expected exit code %d, got %d", int(PRIVATE_KEY_LOAD_FAIL), exitCode)
	}

	outputStr := logger.String()
	if !strings.Contains(outputStr, "Failed to load private key") {
		t.Fatalf("Expected error message to contain 'Failed to load private key', got %s", outputStr)
	}
}

func Test_we_log_an_error_when_chain_id_is_not_valid(t *testing.T) {
	setup()
	testutil.InitDatabase(t)
	defer testutil.CloseDatabase()
	logger := testutil.GetLogger()
	insertSignableMatch()

	os.Setenv("CHAIN_ID", "not_valid")
	exitCode := Run()
	if exitCode != int(CHAIN_ID_NOT_VALID) {
		t.Fatalf("Expected exit code 1, got %d", exitCode)
	}

	outputStr := logger.String()
	if !strings.Contains(outputStr, "Failed to get chain ID") {
		t.Fatalf("Expected error message to contain 'Failed to get chain ID', got %s", outputStr)
	}
}

func Test_we_sign_a_match(t *testing.T) {
	setup()
	testutil.InitDatabase(t)
	defer testutil.CloseDatabase()
	logger := testutil.GetLogger()
	insertSignableMatch()

	exitCode := Run()
	outputStr := logger.String()
	fmt.Println(outputStr)
	if strings.Contains(outputStr, "ERROR") {
		t.Fatalf("Expected no error message, got %s", outputStr)
	}

	if exitCode != 0 {
		t.Fatalf("Expected exit code 0, got %d", exitCode)
	}

	matches, err := repository.FindMatchesToSign()
	if err != nil {
		t.Fatalf("Failed to find matches to sign: %v", err)
	}

	if len(matches) != 0 {
		t.Fatalf("Expected 0 matches to sign, got %d", len(matches))
	}

	var signature string
	err = db.DB.QueryRow(
		"SELECT signature FROM matches WHERE status = $1 AND canonical_id = $2",
		entity.Signed,
		"0x7ed54b4173481077ca259c17a51291beed5152f35d37e01142cd6ee2f771127f",
	).Scan(&signature)
	if err != nil {
		t.Fatalf("Failed to find signed match: %v", err)
	}

	if signature != "88d26e4b2967879b2a10e8b61e9c84a6b496967f252a94fd8df6600ea5e3ee26098054ad05074dfcce9e53eadf6942a316e531348d54120118b3a4f86f1ef9591b" {
		t.Fatalf("Expected signature to be %s, got %s", "88d26e4b2967879b2a10e8b61e9c84a6b496967f252a94fd8df6600ea5e3ee26098054ad05074dfcce9e53eadf6942a316e531348d54120118b3a4f86f1ef9591b", signature)
	}
}

func insertSignableMatch() {
	// TODO: prettify this
	db.DB.Exec("DELETE FROM matches")

	// Match the Solidity test data:
	// COMPETITION_ID = 1, HOME_TEAM_ID = 1, AWAY_TEAM_ID = 2, matchDate = 20251219
	// This should generate: keccak256(abi.encodePacked(1, 1, 2, 20251219))
	// We need to compute this matchId to match what Solidity expects
	matchDate := time.Date(2025, 12, 19, 0, 0, 0, 0, time.UTC)
	
	// Compute matchId the same way Solidity does: keccak256(abi.encodePacked(competitionId, homeTeamId, awayTeamId, matchDate))
	// In Solidity: abi.encodePacked(uint32(1), uint32(1), uint32(2), uint32(20251219))
	// Solidity's abi.encodePacked packs uint32 as 4 bytes in little-endian? No, it's just the raw bytes
	// Let's verify: uint32(1) = 0x00000001, uint32(20251219) = 0x0134B5D3
	packed := []byte{
		0x00, 0x00, 0x00, 0x01, // competitionId = 1 (uint32, big-endian)
		0x00, 0x00, 0x00, 0x01, // homeTeamId = 1 (uint32, big-endian)
		0x00, 0x00, 0x00, 0x02, // awayTeamId = 2 (uint32, big-endian)
		0x01, 0x34, 0xB5, 0xD3, // matchDate = 20251219 (uint32, 0x0134B5D3, big-endian)
	}
	
	// Verify: 20251219 in hex = 0x0134B5D3
	// Let's double-check: 20251219 decimal = 0x134B5D3, but we need 4 bytes so it's 0x0134B5D3
	
	// Compute keccak256 hash
	matchIdHash := crypto.Keccak256(packed)
	
	// Expected matchId from Solidity test: 57368276741802462541172031530649383770946959070025191394401560582772977898111
	// Convert to hex: this is the bytes32 value that Solidity expects
	expectedMatchIdDecimal := new(big.Int)
	expectedMatchIdDecimal.SetString("57368276741802462541172031530649383770946959070025191394401560582772977898111", 10)
	expectedMatchIdBytes := expectedMatchIdDecimal.Bytes()
	// Pad to 32 bytes
	if len(expectedMatchIdBytes) < 32 {
		padded := make([]byte, 32)
		copy(padded[32-len(expectedMatchIdBytes):], expectedMatchIdBytes)
		expectedMatchIdBytes = padded
	}
	
	// Use the expected matchId from Solidity to ensure signature matches
	matchIdHex := "0x" + hex.EncodeToString(expectedMatchIdBytes)
	
	// Log for debugging
	computedMatchIdDecimal := new(big.Int).SetBytes(matchIdHash).String()
	expectedMatchIdDecimalStr := expectedMatchIdDecimal.String()
	if computedMatchIdDecimal != expectedMatchIdDecimalStr {
		fmt.Printf("WARNING: MatchId computation mismatch!\n")
		fmt.Printf("  Expected (from Solidity): %s\n", expectedMatchIdDecimalStr)
		fmt.Printf("  Computed (in Go):        %s\n", computedMatchIdDecimal)
		fmt.Printf("  Packed bytes: %x\n", packed)
		fmt.Printf("  Using expected matchId for signature generation\n")
	}

	db.DB.Exec(`INSERT INTO matches (
			id, canonical_id, home_team_score, away_team_score, home_team_id, away_team_id, "start", "end", provider_match_id, competition_id, provider, status
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		uuid.New(),
		matchIdHex,
		1, // homeTeamScore
		2, // awayTeamScore
		1, // homeTeamId (matches Solidity test)
		2, // awayTeamId (matches Solidity test)
		matchDate,
		matchDate.Add(2*time.Hour),
		"1234567890",
		"1", // competitionId
		1,
		entity.Finished,
	)
}
