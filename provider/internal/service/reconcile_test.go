package service

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"
	"provider/internal/entity"
	"provider/testutil"

	"github.com/google/uuid"
)

//go:embed test_data/football_org/matches/unknown_competition.json
var footballOrgUnknownCompetitionResponse string

//go:embed test_data/apifootball/unknown_competition.json
var apifootballUnknownCompetitionResponse string

//go:embed test_data/football_org/matches/not_finished_match.json
var footballOrgNotFinishedMatchResponse string

//go:embed test_data/apifootball/not_finished_match.json
var apifootballNotFinishedMatchResponse string

//go:embed test_data/football_org/matches/finished_match.json
var footballOrgFinishedMatchResponse string

// apifootballFinishedMatchResponse is defined in sync_apifootball_test.go (same package).

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

func Test_we_log_a_warning_and_increment_tries_when_match_is_not_found(t *testing.T) {
	db, repositories := testutil.InitDB(t)
	defer testutil.CloseDB(db)
	logger := testutil.GetLogger()
	initialTries := 0
	expectedTries := 5

	tests := []struct {
		name       string
		provider   entity.Provider
		statusCode int
		body       string
	}{
		{
			name:       "football_org returns 404 when match not found",
			provider:   entity.FootballOrg,
			statusCode: http.StatusNotFound,
			body:       "",
		},
		{
			name:       "apifootball returns 200 with error body when match not found",
			provider:   entity.APIFootball,
			statusCode: http.StatusOK,
			body:       `{"error":404,"message":"Not found"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := db.Exec("DELETE FROM match_reconciliation")
			if err != nil {
				t.Fatal(err)
			}
			_, err = db.Exec(`
				INSERT INTO match_reconciliation (id, provider_match_id, provider, reconciled_at, tries)
				VALUES ($1, $2, $3, $4, $5)
			`,
				uuid.New().String(),
				"test_match_123",
				tt.provider,
				nil,
				initialTries,
			)
			if err != nil {
				t.Fatal(err)
			}

			mockServer := testutil.CreateServerBuilder().
				WithStatusCode(tt.statusCode).
				WithResponseBody(tt.body).
				Build()
			defer mockServer.Close()

			err = Reconcile(repositories)
			testutil.AssertNoError(t, err)
			testutil.ExpectNumberOfRequests(t, mockServer, uint(expectedTries))
			testutil.AssertMessageGotLogged(t, logger, "failed to fetch match")
			assertTries(t, db, expectedTries)
		})
	}
}

func Test_we_log_an_error_and_increment_tries_when_unable_to_map_response_to_entity(t *testing.T) {
	db, repositories := testutil.InitDB(t)
	defer testutil.CloseDB(db)
	logger := testutil.GetLogger()
	initialTries := 0
	expectedTries := 5

	tests := []struct {
		name       string
		provider   entity.Provider
		statusCode int
		body       string
	}{
		{
			name:       "football_org returns 200 with body that does not map to entity",
			provider:   entity.FootballOrg,
			statusCode: http.StatusOK,
			body:       `{"randomKey": "randomValue"}`,
		},
		{
			name:       "apifootball returns 200 with event that has unmapped team IDs",
			provider:   entity.APIFootball,
			statusCode: http.StatusOK,
			body:       `[{"randomKey": "randomValue"}]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := db.Exec("DELETE FROM match_reconciliation")
			if err != nil {
				t.Fatal(err)
			}
			_, err = db.Exec(`
				INSERT INTO match_reconciliation (id, provider_match_id, provider, reconciled_at, tries)
				VALUES ($1, $2, $3, $4, $5)
			`,
				uuid.New().String(),
				"test_match_123",
				tt.provider,
				nil,
				initialTries,
			)
			if err != nil {
				t.Fatal(err)
			}

			mockServer := testutil.CreateServerBuilder().
				WithStatusCode(tt.statusCode).
				WithResponseBody(tt.body).
				Build()
			defer mockServer.Close()

			err = Reconcile(repositories)
			testutil.AssertNoError(t, err)
			testutil.ExpectNumberOfRequests(t, mockServer, uint(expectedTries))
			testutil.AssertMessageGotLogged(t, logger, "failed to fetch match")
			assertTries(t, db, expectedTries)
		})
	}
}

func Test_we_log_an_error_and_increment_tries_when_unknown_competition_is_found(t *testing.T) {
	db, repositories := testutil.InitDB(t)
	defer testutil.CloseDB(db)
	logger := testutil.GetLogger()
	initialTries := 0
	expectedTries := 5

	tests := []struct {
		name           string
		provider       entity.Provider
		statusCode     int
		body           string
		expectedLogMsg string
	}{
		{
			name:           "football_org returns match with unknown competition ID",
			provider:       entity.FootballOrg,
			statusCode:     http.StatusOK,
			body:           footballOrgUnknownCompetitionResponse,
			expectedLogMsg: "unknown competition ID 999 for match test_match_123",
		},
		{
			name:           "apifootball returns event with unknown league_id",
			provider:       entity.APIFootball,
			statusCode:     http.StatusOK,
			body:           apifootballUnknownCompetitionResponse,
			expectedLogMsg: "unknown league_id 1 for match test_match_123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := db.Exec("DELETE FROM match_reconciliation")
			if err != nil {
				t.Fatal(err)
			}
			_, err = db.Exec(`
				INSERT INTO match_reconciliation (id, provider_match_id, provider, reconciled_at, tries)
				VALUES ($1, $2, $3, $4, $5)
			`,
				uuid.New().String(),
				"test_match_123",
				tt.provider,
				nil,
				initialTries,
			)
			if err != nil {
				t.Fatal(err)
			}

			mockServer := testutil.CreateServerBuilder().
				WithStatusCode(tt.statusCode).
				WithResponseBody(tt.body).
				Build()
			defer mockServer.Close()

			err = Reconcile(repositories)
			testutil.AssertNoError(t, err)
			testutil.ExpectNumberOfRequests(t, mockServer, uint(expectedTries))
			testutil.AssertMessageGotLogged(t, logger, tt.expectedLogMsg)
			assertTries(t, db, expectedTries)
		})
	}
}

func Test_we_log_a_warning_and_increment_tries_when_match_is_not_finished(t *testing.T) {
	db, repositories := testutil.InitDB(t)
	defer testutil.CloseDB(db)
	logger := testutil.GetLogger()
	initialTries := 0
	expectedTries := 5

	tests := []struct {
		name       string
		provider   entity.Provider
		statusCode int
		body       string
	}{
		{
			name:       "football_org returns match not yet finished",
			provider:   entity.FootballOrg,
			statusCode: http.StatusOK,
			body:       footballOrgNotFinishedMatchResponse,
		},
		{
			name:       "apifootball returns event not yet finished",
			provider:   entity.APIFootball,
			statusCode: http.StatusOK,
			body:       apifootballNotFinishedMatchResponse,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := db.Exec("DELETE FROM match_reconciliation")
			if err != nil {
				t.Fatal(err)
			}
			_, err = db.Exec(`
				INSERT INTO match_reconciliation (id, provider_match_id, provider, reconciled_at, tries)
				VALUES ($1, $2, $3, $4, $5)
			`,
				uuid.New().String(),
				"test_match_123",
				tt.provider,
				nil,
				initialTries,
			)
			if err != nil {
				t.Fatal(err)
			}

			mockServer := testutil.CreateServerBuilder().
				WithStatusCode(tt.statusCode).
				WithResponseBody(tt.body).
				Build()
			defer mockServer.Close()

			err = Reconcile(repositories)
			testutil.AssertNoError(t, err)
			testutil.ExpectNumberOfRequests(t, mockServer, uint(expectedTries))
			testutil.AssertMessageGotLogged(t, logger, "match not yet finished, will retry later")
			assertTries(t, db, expectedTries)
		})
	}
}

func Test_we_insert_the_match_and_remove_the_entry_from_the_queue_when_match_is_finished(t *testing.T) {
	db, repositories := testutil.InitDB(t)
	defer testutil.CloseDB(db)
	logger := testutil.GetLogger()
	ctx := context.Background()
	initialTries := 0
	expectedRequests := 1

	tests := []struct {
		name             string
		provider         entity.Provider
		providerMatchID  string
		body             string
		expectedStart    string
		expectedEnd      string
		competition      entity.Competition
		homeTeam         entity.Team
		awayTeam         entity.Team
		homeScore        uint
		awayScore        uint
	}{
		{
			name:            "football_org",
			provider:        entity.FootballOrg,
			providerMatchID: "544391",
			body:            footballOrgFinishedMatchResponse,
			expectedStart:   "2025-12-03 18:00:00",
			expectedEnd:     "2025-12-03 20:00:00",
			competition:     entity.LaLiga,
			homeTeam:        entity.AthleticClub,
			awayTeam:        entity.RealMadrid,
			homeScore:       0,
			awayScore:       3,
		},
		{
			name:            "apifootball",
			provider:        entity.APIFootball,
			providerMatchID: "619300",
			body:            apifootballFinishedMatchResponse,
			expectedStart:   "2026-02-21 15:00:00",
			expectedEnd:     "2026-02-21 17:00:00",
			competition:     entity.Championship,
			homeTeam:        entity.Swansea,
			awayTeam:        entity.BristolCity,
			homeScore:       2,
			awayScore:       1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := db.Exec("DELETE FROM match_reconciliation")
			if err != nil {
				t.Fatal(err)
			}
			_, err = db.Exec(`
				INSERT INTO match_reconciliation (id, provider_match_id, provider, reconciled_at, tries)
				VALUES ($1, $2, $3, $4, $5)
			`,
				uuid.New().String(),
				tt.providerMatchID,
				tt.provider,
				nil,
				initialTries,
			)
			if err != nil {
				t.Fatal(err)
			}

			mockServer := testutil.CreateServerBuilder().
				WithStatusCode(http.StatusOK).
				WithResponseBody(tt.body).
				Build()
			defer mockServer.Close()

			err = Reconcile(repositories)
			testutil.AssertNoError(t, err)
			testutil.ExpectNumberOfRequests(t, mockServer, uint(expectedRequests))
			testutil.AssertMessageGotLogged(t, logger, "reconciled match")
			assertTries(t, db, 0)

			start, err := time.Parse(time.DateTime, tt.expectedStart)
			if err != nil {
				t.Fatal(err)
			}
			end, err := time.Parse(time.DateTime, tt.expectedEnd)
			if err != nil {
				t.Fatal(err)
			}
			expectedForLookup, err := entity.NewMatch(start, tt.provider, tt.providerMatchID, tt.homeTeam, tt.awayTeam, tt.homeScore, tt.awayScore, tt.competition, entity.Finished)
			if err != nil {
				t.Fatal(err)
			}

			actualMatch, err := repositories.Match.FindByCanonicalID(ctx, expectedForLookup.CanonicalID, tt.provider)
			if err != nil {
				t.Fatalf("FindByCanonicalID: %v", err)
			}

			expectedMatch := &entity.Match{
				ID:              actualMatch.ID,
				CanonicalID:     expectedForLookup.CanonicalID,
				Start:           start,
				End:             end,
				Status:          entity.Finished,
				Provider:        tt.provider,
				ProviderMatchID: tt.providerMatchID,
				CompetitionID:   tt.competition,
				HomeTeamID:      tt.homeTeam,
				AwayTeamID:      tt.awayTeam,
				HomeTeamScore:   tt.homeScore,
				AwayTeamScore:   tt.awayScore,
			}
			if !reflect.DeepEqual(expectedMatch, actualMatch) {
				t.Errorf("Expected match %+v, but got %+v", expectedMatch, actualMatch)
			}
		})
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
