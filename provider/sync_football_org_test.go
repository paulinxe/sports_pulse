package main

import (
	"context"
	_ "embed"
	"net/http"
	"provider/db"
	"provider/entity"
	"provider/repository"
	"provider/testutil"
	"reflect"
	"strings"
	"testing"
	"time"
)

//go:embed football_org/test_data_provider/competition_matches/valid_response.json
var successResponse string

//go:embed football_org/test_data_provider/competition_matches/home_team_not_mapped.json
var homeTeamNotMappedResponse string

//go:embed football_org/test_data_provider/competition_matches/away_team_not_mapped.json
var awayTeamNotMappedResponse string

//go:embed football_org/test_data_provider/competition_matches/invalid_match_date.json
var invalidMatchDateResponse string

//go:embed football_org/test_data_provider/competition_matches/finished_match.json
var finishedMatchCompetitionResponse string

//go:embed football_org/test_data_provider/competition_matches/awarded_match.json
var awardedMatchCompetitionResponse string

//go:embed football_org/test_data_provider/competition_matches/stale_and_finished_matches.json
var staleAndFinishedMatchesResponse string

//go:embed football_org/test_data_provider/competition_matches/stale_pending_match.json
var stalePendingMatchResponse string

func Test_we_can_handle_unknown_competition(t *testing.T) {
	err := Sync("football_org", "premier_league")
	if err == nil {
		t.Error("Expected error but got nil", err)
	}

	expectedError := "Competition not handled by football_org provider: 2"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', but got: %s", expectedError, err.Error())
	}
}

func Test_we_skip_the_match_if_home_team_is_not_mapped(t *testing.T) {
	testutil.InitDatabase(t)
	defer testutil.CloseDatabase()

	logger := testutil.GetLogger()
	mockServer := testutil.CreateServerBuilder().
		WithStatusCode(http.StatusOK).
		WithResponseBody(homeTeamNotMappedResponse).
		Build()
	defer mockServer.Close()

	err := Sync("football_org", "la_liga")
	testutil.AssertNoError(t, err)

	outputStr := logger.String()

	// The match with unmapped home team should be skipped due to team mapping error
	if !strings.Contains(outputStr, "Failed to map home team ID (123456), skipping match (654321)") {
		t.Errorf("Expected 'Failed to map home team ID (123456), skipping match (654321)' in output, but got: %s", outputStr)
	}

	// Athletic - Real Madrid match should be saved since it has valid team mappings
	if !testutil.MatchExists(t, "d0d6f75f29b5b1bb1fc3583476993ede1e43a5c07a57e8280159e0a93510c753") {
		t.Errorf("Athletic - Real Madrid match should exist, but it does not")
	}
}

func Test_we_skip_the_match_if_away_team_is_not_mapped(t *testing.T) {
	testutil.InitDatabase(t)
	defer testutil.CloseDatabase()

	logger := testutil.GetLogger()
	mockServer := testutil.CreateServerBuilder().
		WithStatusCode(http.StatusOK).
		WithResponseBody(awayTeamNotMappedResponse).
		Build()
	defer mockServer.Close()

	err := Sync("football_org", "la_liga")
	testutil.AssertNoError(t, err)

	outputStr := logger.String()
	// The match with unmapped away team should be skipped due to team mapping error
	if !strings.Contains(outputStr, "Failed to map away team ID (123456), skipping match (654321)") {
		t.Errorf("Expected 'Failed to map away team ID (123456), skipping match (654321)' in output, but got: %s", outputStr)
	}

	// Athletic - Real Madrid match should be saved since it has valid team mappings
	if !testutil.MatchExists(t, "d0d6f75f29b5b1bb1fc3583476993ede1e43a5c07a57e8280159e0a93510c753") {
		t.Errorf("Athletic - Real Madrid match should exist, but it does not")
	}
}

func Test_we_can_insert_a_match_when_no_matches_exist_for_competition(t *testing.T) {
	testutil.InitDatabase(t)
	defer testutil.CloseDatabase()

	mockServer := testutil.CreateServerBuilder().
		WithStatusCode(http.StatusOK).
		WithResponseBody(successResponse).
		Build()
	defer mockServer.Close()

	err := Sync("football_org", "la_liga")
	testutil.AssertNoError(t, err)

	// TODO: would be nice to use some kind of Clock so we can mock the current time to avoid flakiness
	// When done, we would be able to add a new test to cover this scenario:
	// Now is 23:40
	// The API returns a match that starts 23:50
	// We advance the clock to 00:10
	// We should not advance the sync date to today and keep on yesterday as the match is in play.
	now := time.Now().UTC()
	expectedDateFrom := now.Format("2006-01-02")
	expectedDateTo := now.Add(24 * time.Hour).Format("2006-01-02")

	actualDateFrom := mockServer.GetQueryParam("dateFrom")
	actualDateTo := mockServer.GetQueryParam("dateTo")

	if actualDateFrom != expectedDateFrom {
		t.Errorf("Expected dateFrom to be %s, but got %s", expectedDateFrom, actualDateFrom)
	}

	if actualDateTo != expectedDateTo {
		t.Errorf("Expected dateTo to be %s, but got %s", expectedDateTo, actualDateTo)
	}

	// The match in successResponse is FINISHED, so it should be saved
	expectedMatchStart, _ := time.Parse("2006-01-02 15:04:05", "2025-12-03 18:00:00")
	expectedMatchEnd, _ := time.Parse("2006-01-02 15:04:05", "2025-12-03 20:00:00")

	canonicalID := "d0d6f75f29b5b1bb1fc3583476993ede1e43a5c07a57e8280159e0a93510c753"
	actualMatch, err := repository.FindByCanonicalID(context.Background(), canonicalID, entity.FootballOrg)
	testutil.AssertNoError(t, err)

	if actualMatch == nil {
		t.Fatalf("Expected match to be found, but it is nil")
	}

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

	if !reflect.DeepEqual(*actualMatch, expectedMatch) {
		t.Errorf("Expected match %+v, but got %+v", expectedMatch, actualMatch)
	}
}

func Test_we_insert_a_match_as_finished_when_syncing_a_match_in_final_status(t *testing.T) {
	tests := []struct {
		name              string
		responseBody      string
		providerMatchID   string
		expectedHomeScore uint
		expectedAwayScore uint
		apiStatus         string
	}{
		{
			name:              "FINISHED status maps to Finished",
			responseBody:      finishedMatchCompetitionResponse,
			providerMatchID:   "544214",
			expectedHomeScore: 1,
			expectedAwayScore: 3,
			apiStatus:         "FINISHED",
		},
		{
			name:              "AWARDED status maps to Finished",
			responseBody:      awardedMatchCompetitionResponse,
			providerMatchID:   "544215",
			expectedHomeScore: 0,
			expectedAwayScore: 3,
			apiStatus:         "AWARDED",
		},
	}

	for _, scenario := range tests {
		t.Run(scenario.name, func(t *testing.T) {
			testutil.InitDatabase(t)
			defer testutil.CloseDatabase()

			mockServer := testutil.CreateServerBuilder().
				WithStatusCode(http.StatusOK).
				WithResponseBody(scenario.responseBody).
				Build()
			defer mockServer.Close()

			startTime, _ := time.Parse("2006-01-02 15:04:05", "2025-12-03 18:00:00")
			expectedMatch, err := entity.NewMatch(
				startTime,
				entity.FootballOrg,
				scenario.providerMatchID,
				entity.Girona,
				entity.RayoVallecano,
				scenario.expectedHomeScore,
				scenario.expectedAwayScore,
				entity.LaLiga,
				entity.Finished,
			)
			testutil.AssertNoError(t, err)

			repository.Save(context.Background(), expectedMatch)

			err = Sync("football_org", "la_liga")
			testutil.AssertNoError(t, err)

			actualMatch, err := repository.FindByCanonicalID(context.Background(), expectedMatch.CanonicalID, entity.FootballOrg)
			testutil.AssertNoError(t, err)

			if actualMatch == nil {
				t.Fatalf("Expected match to be found, but it is nil")
			}

			if actualMatch.HomeTeamScore != scenario.expectedHomeScore {
				t.Errorf("Expected match to have home team score %d, but it is %d", scenario.expectedHomeScore, actualMatch.HomeTeamScore)
			}

			if actualMatch.AwayTeamScore != scenario.expectedAwayScore {
				t.Errorf("Expected match to have away team score %d, but it is %d", scenario.expectedAwayScore, actualMatch.AwayTeamScore)
			}

			if actualMatch.Status != entity.Finished {
				t.Errorf("Expected match status to be Finished, but it is %v", actualMatch.Status)
			}

			testutil.ExpectNumberOfRequests(t, mockServer, 1)
		})
	}
}

func Test_no_api_call_is_made_when_last_synced_date_is_in_the_future(t *testing.T) {
	testutil.InitDatabase(t)
	defer testutil.CloseDatabase()

	mockServer := testutil.CreateServerBuilder().
		WithStatusCode(http.StatusOK).
		WithResponseBody("").
		Build()
	defer mockServer.Close()

	futureDate := time.Now().UTC().Add(1 * 24 * time.Hour).Add(1 * time.Minute)
	repository.UpdateLastSyncedDate(context.Background(), entity.LaLiga, entity.FootballOrg, futureDate)

	err := Sync("football_org", "la_liga")
	if err == nil {
		t.Error("Expected error but got nil", err)
	}

	if !strings.Contains(err.Error(), "sync date is in the future") {
		t.Errorf("Expected error to contain 'sync date is in the future', but got: %s", err.Error())
	}

	testutil.ExpectNumberOfRequests(t, mockServer, 0)
}

func Test_sync_state_advances_by_1_day_when_no_matches_are_found(t *testing.T) {
	testutil.InitDatabase(t)
	defer testutil.CloseDatabase()
	logger := testutil.GetLogger()

	emptyMatchesResponse := `{"matches":[]}`
	mockServer := testutil.CreateServerBuilder().
		WithStatusCode(http.StatusOK).
		WithResponseBody(emptyMatchesResponse).
		Build()
	defer mockServer.Close()

	// Set a known sync state
	knownDate := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	repository.UpdateLastSyncedDate(context.Background(), entity.LaLiga, entity.FootballOrg, knownDate)

	err := Sync("football_org", "la_liga")
	testutil.AssertNoError(t, err)

	testutil.ExpectNumberOfRequests(t, mockServer, 1)

	// Verify the sync state was updated to from + 1 day
	// from = knownDate (2025-01-15), so nextSyncAt should be 2025-01-16
	expectedNextSyncAt := knownDate.Add(24 * time.Hour)
	actualLastSyncedDate, err := repository.GetLastSyncedDate(context.Background(), entity.LaLiga, entity.FootballOrg)
	testutil.AssertNoError(t, err)

	if actualLastSyncedDate == nil {
		t.Fatalf("Expected sync state to be updated, but it is nil")
	}

	// Compare dates by formatting as YYYYMMDD
	expectedDateStr := expectedNextSyncAt.Format("20060102")
	actualDateStr := actualLastSyncedDate.Format("20060102")

	if actualDateStr != expectedDateStr {
		t.Errorf("Expected sync state to be %s, but got %s", expectedDateStr, actualDateStr)
	}

	// Verify log message
	outputStr := logger.String()
	if !strings.Contains(outputStr, "All matches finished, advancing sync date by 1 day") {
		t.Errorf("Expected 'All matches finished, advancing sync date by 1 day' in output, but got: %s", outputStr)
	}
}

func Test_sync_state_advances_when_matches_are_found_but_not_in_progress(t *testing.T) {
	testutil.InitDatabase(t)
	defer testutil.CloseDatabase()
	logger := testutil.GetLogger()

	mockServer := testutil.CreateServerBuilder().
		WithStatusCode(http.StatusOK).
		WithResponseBody(successResponse).
		Build()
	defer mockServer.Close()

	// Set a known sync state in the past
	knownDate := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	repository.UpdateLastSyncedDate(context.Background(), entity.LaLiga, entity.FootballOrg, knownDate)

	err := Sync("football_org", "la_liga")
	testutil.AssertNoError(t, err)

	testutil.ExpectNumberOfRequests(t, mockServer, 1)

	// Verify the sync state was advanced by 1 day (matches are TIMED, not in progress, so we advance)
	// from = knownDate (2025-01-15), so next should be 2025-01-16
	expectedNextSyncAt := knownDate.Add(24 * time.Hour)
	actualLastSyncedDate, err := repository.GetLastSyncedDate(context.Background(), entity.LaLiga, entity.FootballOrg)
	testutil.AssertNoError(t, err)

	if actualLastSyncedDate == nil {
		t.Fatalf("Expected sync state to be updated, but it is nil")
	}

	expectedDateStr := expectedNextSyncAt.Format("20060102")
	actualDateStr := actualLastSyncedDate.Format("20060102")

	if actualDateStr != expectedDateStr {
		t.Errorf("Expected sync state to be %s, but got %s", expectedDateStr, actualDateStr)
	}

	outputStr := logger.String()
	if !strings.Contains(outputStr, "All matches finished, advancing sync date by 1 day") {
		t.Errorf("Expected 'All matches finished, advancing sync date by 1 day' in output, but got: %s", outputStr)
	}
}

func Test_first_sync_with_no_matches_stays_on_today(t *testing.T) {
	testutil.InitDatabase(t)
	defer testutil.CloseDatabase()
	logger := testutil.GetLogger()

	emptyMatchesResponse := `{"matches":[]}`
	mockServer := testutil.CreateServerBuilder().
		WithStatusCode(http.StatusOK).
		WithResponseBody(emptyMatchesResponse).
		Build()
	defer mockServer.Close()

	// No sync state exists (first sync)
	err := Sync("football_org", "la_liga")
	testutil.AssertNoError(t, err)

	testutil.ExpectNumberOfRequests(t, mockServer, 1)

	// Verify the sync state was created and set to today (no matches, but we're on today so we stay)
	now := time.Now().UTC()
	expectedNextSyncAt := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	actualLastSyncedDate, err := repository.GetLastSyncedDate(context.Background(), entity.LaLiga, entity.FootballOrg)
	testutil.AssertNoError(t, err)

	if actualLastSyncedDate == nil {
		t.Fatalf("Expected sync state to be created, but it is nil")
	}

	expectedDateStr := expectedNextSyncAt.Format("20060102")
	actualDateStr := actualLastSyncedDate.Format("20060102")

	if actualDateStr != expectedDateStr {
		t.Errorf("Expected sync state to be %s (today), but got %s", expectedDateStr, actualDateStr)
	}

	outputStr := logger.String()
	if !strings.Contains(outputStr, "Staying on today") {
		t.Errorf("Expected 'Staying on today' in output, but got: %s", outputStr)
	}
}

func Test_we_can_handle_invalid_match_date(t *testing.T) {
	testutil.InitDatabase(t)
	defer testutil.CloseDatabase()

	logger := testutil.GetLogger()
	mockServer := testutil.CreateServerBuilder().
		WithStatusCode(http.StatusOK).
		WithResponseBody(invalidMatchDateResponse).
		Build()
	defer mockServer.Close()

	_ = Sync("football_org", "la_liga")

	outputStr := logger.String()
	if !strings.Contains(outputStr, "Failed to parse match date") {
		t.Errorf("Expected 'Failed to parse match date' in output, but got: %s", outputStr)
	}

	if testutil.MatchExists(t, "58a49d03246d65ce3ce64dd7ca690977fe0f2feeccf3403ebe8b95e515599ff8") {
		t.Errorf("Athletic - Real Madrid match should not exist, but it does")
	}
}

func Test_stale_match_moved_to_reconciliation_queue_and_sync_advances(t *testing.T) {
	testutil.InitDatabase(t)
	defer testutil.CloseDatabase()
	logger := testutil.GetLogger()

	mockServer := testutil.CreateServerBuilder().
		WithStatusCode(http.StatusOK).
		WithResponseBody(staleAndFinishedMatchesResponse).
		Build()
	defer mockServer.Close()

	// The test data has matches on 2025-12-03, so sync will query for that day
	lastSyncedDate := time.Date(2025, 12, 3, 0, 0, 0, 0, time.UTC)
	repository.UpdateLastSyncedDate(context.Background(), entity.LaLiga, entity.FootballOrg, lastSyncedDate)

	err := Sync("football_org", "la_liga")
	testutil.AssertNoError(t, err)

	testutil.ExpectNumberOfRequests(t, mockServer, 1)

	// Verify stale match (ID: 999999, status: IN_PLAY, started at 10:00:00Z) is in reconciliation queue
	// This match started more than 6 hours ago
	// TODO: we need to mock the time so we can test this scenario (Clock)
	if !testutil.ReconciliationEntryExists(t, "999999", int(entity.FootballOrg)) {
		t.Errorf("Expected stale match (provider_match_id: 999999) to be in reconciliation queue, but it is not")
	}

	// Verify finished match (ID: 544391) is saved in matches table
	// This match should have canonical_id based on: LaLiga, AthleticClub, RealMadrid, 2025-12-03
	var matchCount int
	err = db.DB.QueryRow("SELECT COUNT(*) FROM matches WHERE provider_match_id = $1 AND provider = $2", "544391", entity.FootballOrg).Scan(&matchCount)
	testutil.AssertNoError(t, err)
	if matchCount == 0 {
		t.Errorf("Expected finished match (provider_match_id: 544391) to be in matches table, but it is not")
	}

	// Verify sync date advanced to next day
	nextDay := time.Date(2025, 12, 4, 0, 0, 0, 0, time.UTC)
	actualLastSyncedDate, err := repository.GetLastSyncedDate(context.Background(), entity.LaLiga, entity.FootballOrg)
	testutil.AssertNoError(t, err)

	if actualLastSyncedDate == nil {
		t.Fatalf("Expected sync state to be updated, but it is nil")
	}

	// Compare dates by formatting as YYYYMMDD
	expectedDateStr := nextDay.Format("20060102")
	actualDateStr := actualLastSyncedDate.Format("20060102")

	if actualDateStr != expectedDateStr {
		t.Errorf("Expected sync state to be %s, but got %s", expectedDateStr, actualDateStr)
	}

	// Verify log message about stale match
	outputStr := logger.String()
	if !strings.Contains(outputStr, "Moved stale match to reconciliation queue") {
		t.Errorf("Expected 'Moved stale match to reconciliation queue' in output, but got: %s", outputStr)
	}

	// Verify log message about advancing sync date
	if !strings.Contains(outputStr, "All matches finished, advancing sync date by 1 day") {
		t.Errorf("Expected 'All matches finished, advancing sync date by 1 day' in output, but got: %s", outputStr)
	}
}

func Test_stale_pending_match_moved_to_reconciliation_queue_and_sync_advances(t *testing.T) {
	testutil.InitDatabase(t)
	defer testutil.CloseDatabase()
	logger := testutil.GetLogger()

	mockServer := testutil.CreateServerBuilder().
		WithStatusCode(http.StatusOK).
		WithResponseBody(stalePendingMatchResponse).
		Build()
	defer mockServer.Close()

	// The test data has matches on 2025-12-03, so sync will query for that day
	lastSyncedDate := time.Date(2025, 12, 3, 0, 0, 0, 0, time.UTC)
	repository.UpdateLastSyncedDate(context.Background(), entity.LaLiga, entity.FootballOrg, lastSyncedDate)

	err := Sync("football_org", "la_liga")
	testutil.AssertNoError(t, err)

	testutil.ExpectNumberOfRequests(t, mockServer, 1)

	// Verify stale match (ID: 888888, status: TIMED/Pending, started at 08:00:00Z) is in reconciliation queue
	// This match should have started more than 6 hours ago but is still Pending
	// TODO: we need to mock the time so we can test this scenario (Clock)
	if !testutil.ReconciliationEntryExists(t, "888888", int(entity.FootballOrg)) {
		t.Errorf("Expected stale pending match (provider_match_id: 888888) to be in reconciliation queue, but it is not")
	}

	// Verify finished match (ID: 544391) is saved in matches table
	var matchCount int
	err = db.DB.QueryRow("SELECT COUNT(*) FROM matches WHERE provider_match_id = $1 AND provider = $2", "544391", entity.FootballOrg).Scan(&matchCount)
	testutil.AssertNoError(t, err)
	if matchCount == 0 {
		t.Errorf("Expected finished match (provider_match_id: 544391) to be in matches table, but it is not")
	}

	// Verify sync date advanced to next day
	nextDay := time.Date(2025, 12, 4, 0, 0, 0, 0, time.UTC)
	actualLastSyncedDate, err := repository.GetLastSyncedDate(context.Background(), entity.LaLiga, entity.FootballOrg)
	testutil.AssertNoError(t, err)

	if actualLastSyncedDate == nil {
		t.Fatalf("Expected sync state to be updated, but it is nil")
	}

	// Compare dates by formatting as YYYYMMDD
	expectedDateStr := nextDay.Format("20060102")
	actualDateStr := actualLastSyncedDate.Format("20060102")

	if actualDateStr != expectedDateStr {
		t.Errorf("Expected sync state to be %s, but got %s", expectedDateStr, actualDateStr)
	}

	// Verify log message about stale match
	outputStr := logger.String()
	if !strings.Contains(outputStr, "Moved stale match to reconciliation queue") {
		t.Errorf("Expected 'Moved stale match to reconciliation queue' in output, but got: %s", outputStr)
	}

	// Verify log message about advancing sync date
	if !strings.Contains(outputStr, "All matches finished, advancing sync date by 1 day") {
		t.Errorf("Expected 'All matches finished, advancing sync date by 1 day' in output, but got: %s", outputStr)
	}
}
