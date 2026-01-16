package football_org

import (
	"context"
	_ "embed"
	"net/http"
	"provider/entity"
	"provider/repository"
	"provider/testutil"
	"reflect"
	"strings"
	"testing"
	"time"
)

//go:embed test_data_provider/competition_matches/valid_response.json
var successResponse string

//go:embed test_data_provider/competition_matches/home_team_not_mapped.json
var homeTeamNotMappedResponse string

//go:embed test_data_provider/competition_matches/away_team_not_mapped.json
var awayTeamNotMappedResponse string

//go:embed test_data_provider/competition_matches/invalid_match_date.json
var invalidMatchDateResponse string

func Test_we_can_handle_unknown_competition(t *testing.T) {
	err := Sync(entity.Competition(0))
	if err == nil {
		t.Error("Expected error but got nil", err)
	}

	expectedError := "unknown competition: 0"
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

	err := Sync(entity.LaLiga)
	testutil.AssertNoError(t, err)

	outputStr := logger.String()

	if !strings.Contains(outputStr, "Failed to map home team ID (123456), skipping match (654321)") {
		t.Errorf("Expected 'Failed to map home team ID (123456), skipping match (654321)' in output, but got: %s", outputStr)
	}

	// We should still create the Athletic - Real Madrid match
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

	err := Sync(entity.LaLiga)
	testutil.AssertNoError(t, err)

	outputStr := logger.String()
	if !strings.Contains(outputStr, "Failed to map away team ID (123456), skipping match (654321)") {
		t.Errorf("Expected 'Failed to map away team ID (123456), skipping match (654321)' in output, but got: %s", outputStr)
	}

	// We should still create the Athletic - Real Madrid match
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

	err := Sync(entity.LaLiga)
	testutil.AssertNoError(t, err)

	// TODO: would be nice to use some kind of Clock so we can mock the current time to avoid flakiness
	now := time.Now()
	expectedDateFrom := now.Format("2006-01-02")
	expectedDateTo := now.Add(3 * 24 * time.Hour).Format("2006-01-02")

	actualDateFrom := mockServer.GetQueryParam("dateFrom")
	actualDateTo := mockServer.GetQueryParam("dateTo")

	if actualDateFrom != expectedDateFrom {
		t.Errorf("Expected dateFrom to be %s, but got %s", expectedDateFrom, actualDateFrom)
	}

	if actualDateTo != expectedDateTo {
		t.Errorf("Expected dateTo to be %s, but got %s", expectedDateTo, actualDateTo)
	}

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
		Status:          entity.Pending,
		Provider:        entity.FootballOrg,
		ProviderMatchID: "544391",
		CompetitionID:   entity.LaLiga,
		HomeTeamID:      entity.AthleticClub,
		AwayTeamID:      entity.RealMadrid,
		HomeTeamScore:   0,
		AwayTeamScore:   0,
	}

	if !reflect.DeepEqual(*actualMatch, expectedMatch) {
		t.Errorf("Expected match %+v, but got %+v", expectedMatch, actualMatch)
	}
}

func Test_we_insert_a_match_as_finished_when_syncing_a_finished_match(t *testing.T) {
	testutil.InitDatabase(t)
	defer testutil.CloseDatabase()

	mockServer := testutil.CreateServerBuilder().
		WithStatusCode(http.StatusOK).
		WithResponseBody(successResponse).
		Build()
	defer mockServer.Close()

	startTime, _ := time.Parse("2006-01-02 15:04:05", "2025-12-03 18:00:00")
	match := entity.NewMatch(
		startTime,
		entity.FootballOrg,
		"544214",
		entity.Girona,
		entity.RayoVallecano,
		1,
		3,
		entity.LaLiga,
		entity.Finished,
	)
	tx, _ := testutil.BeginTransaction(t)
	repository.Save(context.Background(), tx, match)
	tx.Commit()

	err := Sync(entity.LaLiga)
	testutil.AssertNoError(t, err)

	actualMatch, err := repository.FindByCanonicalID(context.Background(), match.CanonicalID, entity.FootballOrg)
	testutil.AssertNoError(t, err)

	if actualMatch == nil {
		t.Fatalf("Expected match to be found, but it is nil")
	}

	if actualMatch.HomeTeamScore != 1 {
		t.Errorf("Expected match to have home team score 1, but it is %d", actualMatch.HomeTeamScore)
	}

	if actualMatch.AwayTeamScore != 3 {
		t.Errorf("Expected match to have away team score 3, but it is %d", actualMatch.AwayTeamScore)
	}

	testutil.ExpectNumberOfRequests(t, mockServer, 1)
}

func Test_no_api_call_is_made_when_last_match_is_already_3_days_in_the_future(t *testing.T) {
	testutil.InitDatabase(t)
	defer testutil.CloseDatabase()
	logger := testutil.GetLogger()

	mockServer := testutil.CreateServerBuilder().
		WithStatusCode(http.StatusOK).
		WithResponseBody("").
		Build()
	defer mockServer.Close()

	futureDate := time.Now().Add(3 * 24 * time.Hour).Add(1 * time.Minute)
	tx, _ := testutil.BeginTransaction(t)
	repository.UpdateLastSyncedDate(context.Background(), tx, entity.LaLiga, entity.FootballOrg, futureDate)
	tx.Commit()

	err := Sync(entity.LaLiga)
	testutil.AssertNoError(t, err)

	testutil.ExpectNumberOfRequests(t, mockServer, 0)
	outputStr := logger.String()
	if !strings.Contains(outputStr, "Sync timestamp is already 3+ days in the future, skipping API call") {
		t.Errorf("Expected 'Sync timestamp is already 3+ days in the future, skipping API call' in output, but got: %s", outputStr)
	}

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
	tx, _ := testutil.BeginTransaction(t)
	repository.UpdateLastSyncedDate(context.Background(), tx, entity.LaLiga, entity.FootballOrg, knownDate)
	tx.Commit()

	err := Sync(entity.LaLiga)
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
	if !strings.Contains(outputStr, "No matches found, advancing sync date by 1 day") {
		t.Errorf("Expected 'No matches found, advancing sync date by 1 day' in output, but got: %s", outputStr)
	}
}

func Test_sync_state_updates_to_end_of_range_when_matches_are_found(t *testing.T) {
	testutil.InitDatabase(t)
	defer testutil.CloseDatabase()
	logger := testutil.GetLogger()

	mockServer := testutil.CreateServerBuilder().
		WithStatusCode(http.StatusOK).
		WithResponseBody(successResponse).
		Build()
	defer mockServer.Close()

	// Set a known sync state
	knownDate := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	tx, _ := testutil.BeginTransaction(t)
	repository.UpdateLastSyncedDate(context.Background(), tx, entity.LaLiga, entity.FootballOrg, knownDate)
	tx.Commit()

	err := Sync(entity.LaLiga)
	testutil.AssertNoError(t, err)

	testutil.ExpectNumberOfRequests(t, mockServer, 1)

	// Verify the sync state was updated to end of range (to = from + 3 days)
	// from = knownDate (2025-01-15), so to should be 2025-01-18
	expectedNextSyncAt := knownDate.Add(3 * 24 * time.Hour)
	actualLastSyncedDate, err := repository.GetLastSyncedDate(context.Background(), entity.LaLiga, entity.FootballOrg)
	testutil.AssertNoError(t, err)

	if actualLastSyncedDate == nil {
		t.Fatalf("Expected sync state to be updated, but it is nil")
	}

	expectedDateStr := expectedNextSyncAt.Format("20060102")
	actualDateStr := actualLastSyncedDate.Format("20060102")

	if actualDateStr != expectedDateStr {
		t.Errorf("Expected sync state to be %s (end of range), but got %s", expectedDateStr, actualDateStr)
	}

	outputStr := logger.String()
	if !strings.Contains(outputStr, "Matches found, updating sync date to end of range") {
		t.Errorf("Expected 'Matches found, updating sync date to end of range' in output, but got: %s", outputStr)
	}
}

func Test_first_sync_with_no_matches_advances_by_1_day_from_today(t *testing.T) {
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
	err := Sync(entity.LaLiga)
	testutil.AssertNoError(t, err)

	testutil.ExpectNumberOfRequests(t, mockServer, 1)

	// Verify the sync state was created and updated to today + 1 day
	now := time.Now()
	expectedNextSyncAt := now.Add(24 * time.Hour)
	actualLastSyncedDate, err := repository.GetLastSyncedDate(context.Background(), entity.LaLiga, entity.FootballOrg)
	testutil.AssertNoError(t, err)

	if actualLastSyncedDate == nil {
		t.Fatalf("Expected sync state to be created, but it is nil")
	}

	expectedDateStr := expectedNextSyncAt.Format("20060102")
	actualDateStr := actualLastSyncedDate.Format("20060102")

	if actualDateStr != expectedDateStr {
		t.Errorf("Expected sync state to be %s (today + 1 day), but got %s", expectedDateStr, actualDateStr)
	}

	outputStr := logger.String()
	if !strings.Contains(outputStr, "No matches found, advancing sync date by 1 day") {
		t.Errorf("Expected 'No matches found, advancing sync date by 1 day' in output, but got: %s", outputStr)
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

	_ = Sync(entity.LaLiga)

	outputStr := logger.String()
	if !strings.Contains(outputStr, "Failed to parse match date") {
		t.Errorf("Expected 'Failed to parse match date' in output, but got: %s", outputStr)
	}

	if testutil.MatchExists(t, "58a49d03246d65ce3ce64dd7ca690977fe0f2feeccf3403ebe8b95e515599ff8") {
		t.Errorf("Athletic - Real Madrid match should not exist, but it does")
	}
}

func Test_we_are_able_to_process_a_match_that_is_already_in_the_database_and_is_in_pending_status(t *testing.T) {
	testutil.InitDatabase(t)
	defer testutil.CloseDatabase()

	mockServer := testutil.CreateServerBuilder().
		WithStatusCode(http.StatusOK).
		WithResponseBody(successResponse).
		Build()
	defer mockServer.Close()

	startTime, _ := time.Parse("2006-01-02 15:04:05", "2025-12-03 18:00:00")
	match := entity.NewMatch(
		startTime,
		entity.FootballOrg,
		"544391",
		entity.AthleticClub,
		entity.RealMadrid,
		1,
		2,
		entity.LaLiga,
		entity.Pending,
	)
	tx, _ := testutil.BeginTransaction(t)
	repository.Save(context.Background(), tx, match)
	tx.Commit()

	err := Sync(entity.LaLiga)
	testutil.AssertNoError(t, err)

	actualMatch, err := repository.FindByCanonicalID(context.Background(), match.CanonicalID, entity.FootballOrg)
	testutil.AssertNoError(t, err)

	if actualMatch == nil {
		t.Fatalf("Expected match to be found, but it is nil")
	}

	if actualMatch.HomeTeamScore != 0 {
		t.Errorf("Expected match to have home team score 0, but it is %d", actualMatch.HomeTeamScore)
	}

	if actualMatch.AwayTeamScore != 0 {
		t.Errorf("Expected match to have away team score 0, but it is %d", actualMatch.AwayTeamScore)
	}

	testutil.ExpectNumberOfRequests(t, mockServer, 1)
}
