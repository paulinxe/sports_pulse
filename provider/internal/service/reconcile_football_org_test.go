package service

import (
	"net/http"
	"testing"
	"provider/testutil"
	"provider/internal/entity"
	"database/sql"
	_ "embed"

	"github.com/google/uuid"
)

//go:embed test_data/matches/unknown_competition.json
var unknownCompetitionResponse string

func Test_we_log_a_warning_and_increment_tries_when_match_is_not_found(t *testing.T) {
	db, repositories := testutil.InitDB(t)
	defer testutil.CloseDB(db)
	logger := testutil.GetLogger()
	initialTries := 0
	expectedTries := 5

	_, _ = db.Exec(`
		INSERT INTO match_reconciliation (id, provider_match_id, provider, reconciled_at, tries)
		VALUES ($1, $2, $3, $4, $5)
	`,
		uuid.New().String(),
		"test_match_123",
		entity.FootballOrg,
		nil,
		initialTries,
	)

	mockServer := testutil.CreateServerBuilder().
		WithStatusCode(http.StatusNotFound).
		WithResponseBody("").
		Build()
	defer mockServer.Close()

	err := Reconcile(repositories)
	testutil.AssertNoError(t, err)
	testutil.ExpectNumberOfRequests(t, mockServer, uint(expectedTries))
	testutil.AssertMessageGotLogged(t, logger, "failed to fetch match")
	assertTries(t, db, expectedTries)
}

func Test_we_log_an_error_and_increment_tries_when_unable_to_map_response_to_entity(t *testing.T) {
	db, repositories := testutil.InitDB(t)
	defer testutil.CloseDB(db)
	logger := testutil.GetLogger()
	initialTries := 0
	expectedTries := 5

	_, err := db.Exec(`
		INSERT INTO match_reconciliation (id, provider_match_id, provider, reconciled_at, tries)
		VALUES ($1, $2, $3, $4, $5)
	`,
		uuid.New().String(),
		"test_match_123",
		entity.FootballOrg,
		nil,
		initialTries,
	)
	testutil.AssertNoError(t, err)

	mockServer := testutil.CreateServerBuilder().
		WithStatusCode(http.StatusOK).
		WithResponseBody(`{"randomKey": "randomValue"}`).
		Build()
	defer mockServer.Close()

	err = Reconcile(repositories)
	testutil.AssertNoError(t, err)
	testutil.ExpectNumberOfRequests(t, mockServer, uint(expectedTries))
	testutil.AssertMessageGotLogged(t, logger, "failed to fetch match")
	assertTries(t, db, expectedTries)
}

func Test_we_log_an_error_and_increment_tries_when_unknown_competition_is_found(t *testing.T) {
	db, repositories := testutil.InitDB(t)
	defer testutil.CloseDB(db)
	logger := testutil.GetLogger()
	initialTries := 0
	expectedTries := 5

	_, err := db.Exec(`
		INSERT INTO match_reconciliation (id, provider_match_id, provider, reconciled_at, tries)
		VALUES ($1, $2, $3, $4, $5)
	`,
		uuid.New().String(),
		"test_match_123",
		entity.FootballOrg,
		nil,
		initialTries,
	)
	testutil.AssertNoError(t, err)

	mockServer := testutil.CreateServerBuilder().
		WithStatusCode(http.StatusOK).
		WithResponseBody(unknownCompetitionResponse).
		Build()
	defer mockServer.Close()

	err = Reconcile(repositories)
	testutil.AssertNoError(t, err)
	testutil.ExpectNumberOfRequests(t, mockServer, uint(expectedTries))
	testutil.AssertMessageGotLogged(t, logger, "unknown competition ID 999 for match test_match_123")
	assertTries(t, db, expectedTries)
}

func assertTries(t *testing.T, db *sql.DB, expectedTries int) {
	t.Helper()
	var tries int

	err := db.QueryRow("SELECT tries FROM match_reconciliation LIMIT 1").Scan(&tries)
	testutil.AssertNoError(t, err)

	if tries != expectedTries {
		t.Errorf("Expected tries to be %d, but got %d", expectedTries, tries)
	}
}
