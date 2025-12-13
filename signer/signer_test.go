package main

import (
	"testing"
	"signer/repository"
	"signer/testutil"
	"signer/db"
	"signer/entity"
	"github.com/google/uuid"
	"os"
	"time"
	"strings"
)

func Test_no_errors_when_nothing_to_sign(t *testing.T) {
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
	testutil.InitDatabase(t)
	defer testutil.CloseDatabase()

	logger := testutil.GetLogger()

	insertSignableMatch(t)

	os.Setenv("PRIVATE_KEY_FILE", "not_found.pem")
	exitCode := Run()
	if exitCode != 1 {
		t.Fatalf("Expected exit code 1, got %d", exitCode)
	}

	outputStr := logger.String()
	if !strings.Contains(outputStr, "Failed to load private key") {
		t.Fatalf("Expected error message to contain 'Failed to load private key', got %s", outputStr)
	}
}

func insertSignableMatch(t *testing.T) {
	testutil.InitDatabase(t)
	defer testutil.CloseDatabase()

	// TODO: prettify this
	db.DB.Exec(`INSERT INTO matches (
			id, canonical_id, home_team_score, away_team_score, home_team_id, away_team_id, "start", "end", provider_match_id, competition_id, provider, status
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		uuid.New(),
		"\x58a49d03246d65ce3ce64dd7ca690977fe0f2feeccf3403ebe8b95e515599ff8",
		1,
		2,
		3,
		4,
		time.Now(),
		time.Now().Add(2 * time.Hour),
		"1234567890",
		"1",
		1,
		entity.Finished,
	)
}