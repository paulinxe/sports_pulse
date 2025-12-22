package main

import (
	"fmt"
	"os"
	"signer/db"
	"signer/entity"
	"signer/repository"
	"signer/testutil"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func setup() {
	os.Setenv("CHAIN_ID", "31337")
	os.Setenv("SIGNER_PRIVATE_KEY", "0x4ba521e286bca3aa5fe1a8a8cf38017246b15fd2a4d9c79f1576ca82b9244279")
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

func Test_we_log_an_error_when_private_key_is_not_valid(t *testing.T) {
	setup()
	testutil.InitDatabase(t)
	defer testutil.CloseDatabase()
	logger := testutil.GetLogger()
	insertSignableMatch()

	os.Setenv("SIGNER_PRIVATE_KEY", "0x69")

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

	if signature != "4f0fa54d6dd9629d5f1d6b0f17236f4f9f009b72be6e77bdc56a4d0d891c0c076f6c36472f7b667d5f63895424a19a19bc56f264e49699c58bb07ec0868440081c" {
		t.Fatalf("Expected signature to be %s, got %s", "4f0fa54d6dd9629d5f1d6b0f17236f4f9f009b72be6e77bdc56a4d0d891c0c076f6c36472f7b667d5f63895424a19a19bc56f264e49699c58bb07ec0868440081c", signature)
	}
}

func insertSignableMatch() {
	// TODO: prettify this
	db.DB.Exec("DELETE FROM matches")

	// Match the Solidity test data:
	// COMPETITION_ID = 1, HOME_TEAM_ID = 1, AWAY_TEAM_ID = 2, matchDate = 20251219
	// This should generate: keccak256(abi.encodePacked(1, 1, 2, 20251219))
	matchDate := time.Date(2025, 12, 19, 0, 0, 0, 0, time.UTC)
	matchIdHex := "0x7ed54b4173481077ca259c17a51291beed5152f35d37e01142cd6ee2f771127f"
	homeTeamScore := 1
	awayTeamScore := 2
	homeTeamId := 1
	awayTeamId := 2

	db.DB.Exec(`INSERT INTO matches (
			id, canonical_id, home_team_score, away_team_score, home_team_id, away_team_id, "start", "end", provider_match_id, competition_id, provider, status
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		uuid.New(),
		matchIdHex,
		homeTeamScore,
		awayTeamScore,
		homeTeamId,
		awayTeamId,
		matchDate,
		matchDate.Add(2*time.Hour),
		"1234567890",
		"1", // competitionId
		1,
		entity.Finished,
	)
}
