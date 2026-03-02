package service

import (
	"context"
	_ "embed"
	"net/http"
	"provider/internal/entity"
	"provider/testutil"
	"reflect"
	"testing"
	"time"
)

//go:embed test_data/apifootball/valid_response.json
var apifootballSuccessResponse string

//go:embed test_data/apifootball/home_team_not_mapped.json
var apifootballHomeTeamNotMappedResponse string

//go:embed test_data/apifootball/away_team_not_mapped.json
var apifootballAwayTeamNotMappedResponse string

//go:embed test_data/apifootball/invalid_match_date.json
var apifootballInvalidMatchDateResponse string

//go:embed test_data/apifootball/finished_match.json
var apifootballFinishedMatchResponse string

//go:embed test_data/apifootball/stale_and_finished_matches.json
var apifootballStaleAndFinishedMatchesResponse string

//go:embed test_data/apifootball/stale_pending_match.json
var apifootballStalePendingMatchResponse string

func Test_apifootball_we_can_handle_unknown_competition(t *testing.T) {
	db, repositories := testutil.InitDB(t)
	defer testutil.CloseDB(db)
	err := Sync(repositories, "apifootball", "la_liga", &SystemClock{})
	if err == nil {
		t.Error("Expected error but got nil", err)
	}

	expectedError := "competition not handled by apifootball provider: 1"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', but got: %s", expectedError, err.Error())
	}
}

func Test_apifootball_we_skip_the_match_if_home_team_is_not_mapped(t *testing.T) {
	db, repositories := testutil.InitDB(t)
	defer testutil.CloseDB(db)
	logger := testutil.GetLogger()
	mockServer := testutil.CreateServerBuilder().
		WithStatusCode(http.StatusOK).
		WithResponseBody(apifootballHomeTeamNotMappedResponse).
		Build()
	defer mockServer.Close()

	err := Sync(repositories, "apifootball", "championship", &SystemClock{})
	testutil.AssertNoError(t, err)
	testutil.AssertMessageGotLogged(t, logger, "unmapped home team ID 9999")

	// We still make sure we can save valid matches
	if !testutil.MatchExists(t, db, "1a9dccd8fa8ed283363f18886e457fe3ba1f2f756f11ab502cb4c67961da16ec") {
		t.Errorf("Swansea - Bristol City match should exist, but it does not")
	}
}

func Test_apifootball_we_skip_the_match_if_away_team_is_not_mapped(t *testing.T) {
	db, repositories := testutil.InitDB(t)
	defer testutil.CloseDB(db)
	logger := testutil.GetLogger()
	mockServer := testutil.CreateServerBuilder().
		WithStatusCode(http.StatusOK).
		WithResponseBody(apifootballAwayTeamNotMappedResponse).
		Build()
	defer mockServer.Close()

	err := Sync(repositories, "apifootball", "championship", &SystemClock{})
	testutil.AssertNoError(t, err)
	testutil.AssertMessageGotLogged(t, logger, "unmapped away team ID 9999")

	// We still make sure we can save valid matches
	if !testutil.MatchExists(t, db, "1a9dccd8fa8ed283363f18886e457fe3ba1f2f756f11ab502cb4c67961da16ec") {
		t.Errorf("Swansea - Bristol City match should exist, but it does not")
	}
}

func Test_apifootball_we_insert_a_match_as_finished_when_syncing_a_match_in_final_status(t *testing.T) {
	db, repositories := testutil.InitDB(t)
	defer testutil.CloseDB(db)

	mockTime := time.Date(2026, 2, 21, 12, 0, 0, 0, time.UTC)
	clock := &mockClock{now: mockTime}

	mockServer := testutil.CreateServerBuilder().
		WithStatusCode(http.StatusOK).
		WithResponseBody(apifootballFinishedMatchResponse).
		Build()
	defer mockServer.Close()

	expectedStart, _ := time.Parse("2006-01-02 15:04", "2026-02-21 15:00")
	expectedMatch, err := entity.NewMatch(
		expectedStart,
		entity.APIFootball,
		"619300",
		entity.Swansea,
		entity.BristolCity,
		2,
		1,
		entity.Championship,
		entity.Finished,
	)
	testutil.AssertNoError(t, err)

	_ = repositories.Match.Save(context.Background(), *expectedMatch)

	err = Sync(repositories, "apifootball", "championship", clock)
	testutil.AssertNoError(t, err)

	actualMatch, err := repositories.Match.FindByCanonicalID(context.Background(), expectedMatch.CanonicalID, entity.APIFootball)
	testutil.AssertNoError(t, err)

	if !reflect.DeepEqual(*actualMatch, *expectedMatch) {
		t.Errorf("Expected match %+v, but got %+v", *expectedMatch, *actualMatch)
	}

	testutil.ExpectNumberOfRequests(t, mockServer, 1)
}

func Test_apifootball_today_is_used_as_query_date_when_last_synced_date_is_in_the_future(t *testing.T) {
	db, repositories := testutil.InitDB(t)
	defer testutil.CloseDB(db)

	mockServer := testutil.CreateServerBuilder().
		WithStatusCode(http.StatusOK).
		WithResponseBody(`[]`).
		Build()
	defer mockServer.Close()

	futureDate := time.Now().UTC().Add(1 * 24 * time.Hour).Add(1 * time.Minute)
	_ = repositories.SyncState.UpdateLastSyncedDate(context.Background(), entity.Championship, entity.APIFootball, futureDate)

	err := Sync(repositories, "apifootball", "championship", &SystemClock{})
	testutil.AssertNoError(t, err)
	testutil.ExpectNumberOfRequests(t, mockServer, 1)

	actualLastSyncedDate, err := repositories.SyncState.GetLastSyncedDate(context.Background(), entity.Championship, entity.APIFootball)
	testutil.AssertNoError(t, err)

	if actualLastSyncedDate == nil {
		t.Fatalf("Expected sync state to be updated, but it is nil")
	}

	expectedDateStr := time.Now().UTC().Format("20060102")
	actualDateStr := actualLastSyncedDate.Format("20060102")

	if actualDateStr != expectedDateStr {
		t.Errorf("Expected sync state to be %s, but got %s", expectedDateStr, actualDateStr)
	}
}

func Test_apifootball_sync_state_advances_by_1_day_when_no_matches_are_found(t *testing.T) {
	db, repositories := testutil.InitDB(t)
	defer testutil.CloseDB(db)
	logger := testutil.GetLogger()

	mockServer := testutil.CreateServerBuilder().
		WithStatusCode(http.StatusOK).
		WithResponseBody(`[]`).
		Build()
	defer mockServer.Close()

	knownDate := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	_ = repositories.SyncState.UpdateLastSyncedDate(context.Background(), entity.Championship, entity.APIFootball, knownDate)

	err := Sync(repositories, "apifootball", "championship", &SystemClock{})
	testutil.AssertNoError(t, err)

	testutil.ExpectNumberOfRequests(t, mockServer, 1)

	expectedNextSyncAt := knownDate.Add(24 * time.Hour)
	actualLastSyncedDate, err := repositories.SyncState.GetLastSyncedDate(context.Background(), entity.Championship, entity.APIFootball)
	testutil.AssertNoError(t, err)

	if actualLastSyncedDate == nil {
		t.Fatalf("Expected sync state to be updated, but it is nil")
	}

	expectedDateStr := expectedNextSyncAt.Format("20060102")
	actualDateStr := actualLastSyncedDate.Format("20060102")

	if actualDateStr != expectedDateStr {
		t.Errorf("Expected sync state to be %s, but got %s", expectedDateStr, actualDateStr)
	}

	testutil.AssertMessageGotLogged(t, logger, "All matches finished, advancing sync date by 1 day")
}

func Test_apifootball_sync_state_advances_when_matches_are_found_but_not_in_progress(t *testing.T) {
	db, repositories := testutil.InitDB(t)
	defer testutil.CloseDB(db)
	logger := testutil.GetLogger()

	mockTime := time.Date(2026, 2, 21, 12, 0, 0, 0, time.UTC)
	clock := &mockClock{now: mockTime}

	mockServer := testutil.CreateServerBuilder().
		WithStatusCode(http.StatusOK).
		WithResponseBody(apifootballSuccessResponse).
		Build()
	defer mockServer.Close()

	knownDate := time.Date(2026, 2, 20, 12, 0, 0, 0, time.UTC)
	_ = repositories.SyncState.UpdateLastSyncedDate(context.Background(), entity.Championship, entity.APIFootball, knownDate)

	err := Sync(repositories, "apifootball", "championship", clock)
	testutil.AssertNoError(t, err)

	testutil.ExpectNumberOfRequests(t, mockServer, 1)

	expectedNextSyncAt := knownDate.Add(24 * time.Hour)
	actualLastSyncedDate, err := repositories.SyncState.GetLastSyncedDate(context.Background(), entity.Championship, entity.APIFootball)
	testutil.AssertNoError(t, err)

	if actualLastSyncedDate == nil {
		t.Fatalf("Expected sync state to be updated, but it is nil")
	}

	expectedDateStr := expectedNextSyncAt.Format("20060102")
	actualDateStr := actualLastSyncedDate.Format("20060102")

	if actualDateStr != expectedDateStr {
		t.Errorf("Expected sync state to be %s, but got %s", expectedDateStr, actualDateStr)
	}

	testutil.AssertMessageGotLogged(t, logger, "All matches finished, advancing sync date by 1 day")
}

func Test_apifootball_first_sync_with_no_matches_stays_on_today(t *testing.T) {
	db, repositories := testutil.InitDB(t)
	defer testutil.CloseDB(db)
	logger := testutil.GetLogger()

	mockServer := testutil.CreateServerBuilder().
		WithStatusCode(http.StatusOK).
		WithResponseBody(`[]`).
		Build()
	defer mockServer.Close()

	err := Sync(repositories, "apifootball", "championship", &SystemClock{})
	testutil.AssertNoError(t, err)

	testutil.ExpectNumberOfRequests(t, mockServer, 1)

	now := time.Now().UTC()
	expectedNextSyncAt := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	actualLastSyncedDate, err := repositories.SyncState.GetLastSyncedDate(context.Background(), entity.Championship, entity.APIFootball)
	testutil.AssertNoError(t, err)

	if actualLastSyncedDate == nil {
		t.Fatalf("Expected sync state to be created, but it is nil")
	}

	expectedDateStr := expectedNextSyncAt.Format("20060102")
	actualDateStr := actualLastSyncedDate.Format("20060102")

	if actualDateStr != expectedDateStr {
		t.Errorf("Expected sync state to be %s (today), but got %s", expectedDateStr, actualDateStr)
	}

	testutil.AssertMessageGotLogged(t, logger, "Staying on today")
}

func Test_apifootball_we_can_handle_invalid_match_date(t *testing.T) {
	db, repositories := testutil.InitDB(t)
	defer testutil.CloseDB(db)
	logger := testutil.GetLogger()

	mockTime := time.Date(2026, 2, 21, 12, 0, 0, 0, time.UTC)
	clock := &mockClock{now: mockTime}

	mockServer := testutil.CreateServerBuilder().
		WithStatusCode(http.StatusOK).
		WithResponseBody(apifootballInvalidMatchDateResponse).
		Build()
	defer mockServer.Close()

	err := Sync(repositories, "apifootball", "championship", clock)
	testutil.AssertNoError(t, err)
	testutil.AssertMessageGotLogged(t, logger, "parse match_date/time")

	var matchCount int
	err = db.QueryRow("SELECT COUNT(*) FROM matches WHERE provider_match_id = $1 AND provider = $2", "619299", entity.APIFootball).Scan(&matchCount)
	testutil.AssertNoError(t, err)
	if matchCount > 0 {
		t.Errorf("Match with invalid date (provider_match_id: 619299) should not exist, but it does")
	}
}

func Test_apifootball_stale_match_moved_to_reconciliation_queue_and_sync_stays_on_today(t *testing.T) {
	db, repositories := testutil.InitDB(t)
	defer testutil.CloseDB(db)
	logger := testutil.GetLogger()

	mockTime := time.Date(2026, 2, 21, 14, 1, 0, 0, time.UTC)
	clock := &mockClock{now: mockTime}

	mockServer := testutil.CreateServerBuilder().
		WithStatusCode(http.StatusOK).
		WithResponseBody(apifootballStaleAndFinishedMatchesResponse).
		Build()
	defer mockServer.Close()

	lastSyncedDate := time.Date(2026, 2, 21, 0, 0, 0, 0, time.UTC)
	_ = repositories.SyncState.UpdateLastSyncedDate(context.Background(), entity.Championship, entity.APIFootball, lastSyncedDate)

	err := Sync(repositories, "apifootball", "championship", clock)
	testutil.AssertNoError(t, err)

	testutil.ExpectNumberOfRequests(t, mockServer, 1)

	if !testutil.ReconciliationEntryExists(t, db, "999999", int(entity.APIFootball)) {
		t.Errorf("Expected stale match (provider_match_id: 999999) to be in reconciliation queue, but it is not")
	}

	var matchCount int
	err = db.QueryRow("SELECT COUNT(*) FROM matches WHERE provider_match_id = $1 AND provider = $2", "619298", entity.APIFootball).Scan(&matchCount)
	testutil.AssertNoError(t, err)
	if matchCount == 0 {
		t.Errorf("Expected finished match (provider_match_id: 619298) to be in matches table, but it is not")
	}

	actualLastSyncedDate, err := repositories.SyncState.GetLastSyncedDate(context.Background(), entity.Championship, entity.APIFootball)
	testutil.AssertNoError(t, err)

	if actualLastSyncedDate == nil {
		t.Fatalf("Expected sync state to be updated, but it is nil")
	}

	expectedDateStr := clock.Now().Format("20060102")
	actualDateStr := actualLastSyncedDate.Format("20060102")

	if actualDateStr != expectedDateStr {
		t.Errorf("Expected sync state to be %s, but got %s", expectedDateStr, actualDateStr)
	}

	testutil.AssertMessageGotLogged(t, logger, "Moved stale match to reconciliation queue")
	testutil.AssertMessageGotLogged(t, logger, "Staying on today")
}

func Test_apifootball_stale_match_moved_to_reconciliation_queue_and_sync_advances(t *testing.T) {
	db, repositories := testutil.InitDB(t)
	defer testutil.CloseDB(db)
	logger := testutil.GetLogger()

	// Clock is 2026-02-22 so "today" is 1 day in the future after our last synced date
	mockTime := time.Date(2026, 2, 22, 12, 0, 0, 0, time.UTC)
	clock := &mockClock{now: mockTime}

	lastSyncedDate := time.Date(2026, 2, 21, 0, 0, 0, 0, time.UTC)
	_ = repositories.SyncState.UpdateLastSyncedDate(context.Background(), entity.Championship, entity.APIFootball, lastSyncedDate)

	mockServer := testutil.CreateServerBuilder().
		WithStatusCode(http.StatusOK).
		WithResponseBody(apifootballStalePendingMatchResponse).
		Build()
	defer mockServer.Close()

	err := Sync(repositories, "apifootball", "championship", clock)
	testutil.AssertNoError(t, err)

	testutil.ExpectNumberOfRequests(t, mockServer, 1)

	if !testutil.ReconciliationEntryExists(t, db, "888888", int(entity.APIFootball)) {
		t.Errorf("Expected stale pending match (provider_match_id: 888888) to be in reconciliation queue, but it is not")
	}

	var matchCount int
	err = db.QueryRow("SELECT COUNT(*) FROM matches WHERE provider_match_id = $1 AND provider = $2", "619298", entity.APIFootball).Scan(&matchCount)
	testutil.AssertNoError(t, err)
	if matchCount == 0 {
		t.Errorf("Expected finished match (provider_match_id: 619298) to be in matches table, but it is not")
	}

	nextDay := time.Date(2026, 2, 22, 0, 0, 0, 0, time.UTC)
	actualLastSyncedDate, err := repositories.SyncState.GetLastSyncedDate(context.Background(), entity.Championship, entity.APIFootball)
	testutil.AssertNoError(t, err)

	if actualLastSyncedDate == nil {
		t.Fatalf("Expected sync state to be updated, but it is nil")
	}

	expectedDateStr := nextDay.Format("20060102")
	actualDateStr := actualLastSyncedDate.Format("20060102")

	if actualDateStr != expectedDateStr {
		t.Errorf("Expected sync state to be %s, but got %s", expectedDateStr, actualDateStr)
	}

	testutil.AssertMessageGotLogged(t, logger, "Moved stale match to reconciliation queue")
	testutil.AssertMessageGotLogged(t, logger, "All matches finished, advancing sync date by 1 day")
}
