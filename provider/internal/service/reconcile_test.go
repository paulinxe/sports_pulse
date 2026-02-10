package service

import (
	"strings"
	"testing"
	"provider/testutil"

	"github.com/google/uuid"
)

func Test_process_ends_successfully_when_no_reconciliable_matches_are_found(t *testing.T) {
	db, repositories := testutil.InitDB(t)
	defer testutil.CloseDB(db)

	err := Reconcile(repositories)
	testutil.AssertNoError(t, err)
}

func Test_we_log_an_error_and_increment_tries_when_unable_to_map_provider(t *testing.T) {
	db, repositories := testutil.InitDB(t)
	defer testutil.CloseDB(db)
	logger := testutil.GetLogger()
	initialTries := 0
	expectedTries := 5 // This is maximum number of tries allow

	_, err := db.Exec(`
		INSERT INTO match_reconciliation (id, provider_match_id, provider, reconciled_at, tries)
		VALUES ($1, $2, $3, $4, $5)
	`,
		uuid.New().String(),
		"test_match_123",
		69,
		nil,
		initialTries,
	)
	if err != nil {
		t.Error("Expected no error but got", err)
	}

	err = Reconcile(repositories)
	// As we iterate, the commands needs to finish successfully
	testutil.AssertNoError(t, err)

	outputStr := logger.String()
	if !strings.Contains(outputStr, "unable to get provider for reconciliation. manual intervention required.") {
		t.Errorf("Expected log 'unable to get provider for reconciliation. manual intervention required.', but got: %s", outputStr)
	}

	var tries int
	err = db.QueryRow("SELECT tries FROM match_reconciliation LIMIT 1").Scan(&tries)
	testutil.AssertNoError(t, err)
	if tries != expectedTries {
		t.Errorf("Expected tries to be %d, but got %d", expectedTries, tries)
	}
}