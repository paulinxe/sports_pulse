package service

import (
	"context"
	"net/http"
	"testing"
	"provider/testutil"
	"provider/internal/entity"
	"database/sql"
	"errors"
	_ "embed"
	"reflect"
	"time"

	"github.com/google/uuid"
)

//go:embed test_data/football_org/matches/unknown_competition.json
var unknownCompetitionResponse string

//go:embed test_data/football_org/matches/not_finished_match.json
var notFinishedMatchResponse string

//go:embed test_data/football_org/matches/finished_match.json
var finishedMatchResponse string

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

func Test_we_log_a_warning_and_increment_tries_when_match_is_not_finished(t *testing.T) {
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
		WithStatusCode(http.StatusOK).
		WithResponseBody(notFinishedMatchResponse).
		Build()
	defer mockServer.Close()

	err := Reconcile(repositories)
	testutil.AssertNoError(t, err)
	testutil.ExpectNumberOfRequests(t, mockServer, uint(expectedTries))
	testutil.AssertMessageGotLogged(t, logger, "match not yet finished, will retry later")
	assertTries(t, db, expectedTries)
}

func Test_we_insert_the_match_and_remove_the_entry_from_the_queue_when_match_is_finished(t *testing.T) {
	db, repositories := testutil.InitDB(t)
	defer testutil.CloseDB(db)
	logger := testutil.GetLogger()
	initialTries := 0
	expectedTries := 1
	
	_, _ = db.Exec(`
		INSERT INTO match_reconciliation (id, provider_match_id, provider, reconciled_at, tries)
		VALUES ($1, $2, $3, $4, $5)
	`,
		uuid.New().String(),
		"544391",
		entity.FootballOrg,
		nil,
		initialTries,
	)
	mockServer := testutil.CreateServerBuilder().
		WithStatusCode(http.StatusOK).
		WithResponseBody(finishedMatchResponse).
		Build()
	defer mockServer.Close()

	err := Reconcile(repositories)
	testutil.AssertNoError(t, err)
	testutil.ExpectNumberOfRequests(t, mockServer, uint(expectedTries))
	testutil.AssertMessageGotLogged(t, logger, "reconciled match")
	assertTries(t, db, 0) // After the match is reconciled, the entry is removed from the queue

	canonicalID := "d0d6f75f29b5b1bb1fc3583476993ede1e43a5c07a57e8280159e0a93510c753"
	actualMatch, _ := repositories.Match.FindByCanonicalID(context.Background(), canonicalID, entity.FootballOrg)
	expectedMatchStart, _ := time.Parse("2006-01-02 15:04:05", "2025-12-03 18:00:00")
	expectedMatchEnd, _ := time.Parse("2006-01-02 15:04:05", "2025-12-03 20:00:00")
	expectedMatch := entity.Match{
		ID:              actualMatch.ID, // Small hack to be able to compare the matches
		CanonicalID:     canonicalID,
		Start:           expectedMatchStart,
		End:             expectedMatchEnd,
		Status:          entity.Finished,
		Provider:        entity.FootballOrg,
		ProviderMatchID: "544391",
		CompetitionID:   entity.LaLiga,
		HomeTeamID:      entity.AthleticClub,
		AwayTeamID:      entity.RealMadrid,
		HomeTeamScore:   0,
		AwayTeamScore:   3,
	}

	if !reflect.DeepEqual(&expectedMatch, actualMatch) {
		t.Errorf("Expected match %+v, but got %+v", expectedMatch, actualMatch)
	}
}

func assertTries(t *testing.T, db *sql.DB, expectedTries int) {
	t.Helper()
	var tries int

	err := db.QueryRow("SELECT tries FROM match_reconciliation LIMIT 1").Scan(&tries)
	if expectedTries == 0 && errors.Is(err, sql.ErrNoRows) {
		return
	}

	testutil.AssertNoError(t, err)

	if tries != expectedTries {
		t.Errorf("Expected tries to be %d, but got %d", expectedTries, tries)
	}
}
