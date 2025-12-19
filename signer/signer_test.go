package main

import (
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
	os.Setenv("PRIVATE_KEY_FILE", "private_test_key.key")
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
	if exitCode != 0 {
		t.Fatalf("Expected exit code 0, got %d", exitCode)
	}

	outputStr := logger.String()
	if strings.Contains(outputStr, "ERROR") {
		t.Fatalf("Expected no error message, got %s", outputStr)
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
		"0x58a49d03246d65ce3ce64dd7ca690977fe0f2feeccf3403ebe8b95e515599ff8",
	).Scan(&signature)
	if err != nil {
		t.Fatalf("Failed to find signed match: %v", err)
	}

	if signature != "66c24c18bcd615e63d0b828fd5f0a088e06e9985a61ccc272f05eb6f377ff8444f52347cbfebecabb25e5423289f668dbc25aac4bd98f6bd3445190e9ebed39200" {
		t.Fatalf("Expected signature to be %s, got %s", "66c24c18bcd615e63d0b828fd5f0a088e06e9985a61ccc272f05eb6f377ff8444f52347cbfebecabb25e5423289f668dbc25aac4bd98f6bd3445190e9ebed39200", signature)
	}
}

func insertSignableMatch() {
	// TODO: prettify this
	db.DB.Exec("DELETE FROM matches")

	db.DB.Exec(`INSERT INTO matches (
			id, canonical_id, home_team_score, away_team_score, home_team_id, away_team_id, "start", "end", provider_match_id, competition_id, provider, status
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		uuid.New(),
		"0x58a49d03246d65ce3ce64dd7ca690977fe0f2feeccf3403ebe8b95e515599ff8",
		1,
		2,
		3,
		4,
		time.Now(),
		time.Now().Add(2*time.Hour),
		"1234567890",
		"1",
		1,
		entity.Finished,
	)
}
