package football_org

import (
	_ "embed"
	"net/http"
	"provider/entity"
	"provider/repository"
	"provider/testutil"
	"testing"
	"time"
)

//go:embed test_data_provider/matches/not_finished_match.json
var notFinishedMatchResponse string

//go:embed test_data_provider/matches/finished_match.json
var finishedMatchResponse string

func Test_no_errors_when_nothing_to_reconcile(t *testing.T) {
	testutil.InitDatabase(t)
	defer testutil.CloseDatabase()

	mockServer := testutil.CreateServer(http.StatusOK, "")
	defer mockServer.Close()

	err := Reconcile()
	if err != nil {
		t.Errorf("Expected no error but got: %v", t)
	}

	testutil.ExpectNumberOfRequests(t, mockServer, 0)
}

func Test_we_ignore_matches_that_are_pending_for_more_than_24_hours(t *testing.T) {
	testutil.InitDatabase(t)
	defer testutil.CloseDatabase()

	mockServer := testutil.CreateServer(http.StatusOK, "")
	defer mockServer.Close()

	// 26 hours ago is the start so the end is 24 hours ago
	startTime, _ := time.Parse("2006-01-02 15:04:05", time.Now().Add(-26*time.Hour).Format("2006-01-02 15:04:05"))
	createMatch(t, startTime)

	err := Reconcile()
	if err != nil {
		t.Errorf("Expected no error but got: %v", err)
	}

	testutil.ExpectNumberOfRequests(t, mockServer, 0)
}

func Test_we_dont_update_the_match_if_the_status_is_not_finished(t *testing.T) {
	testutil.InitDatabase(t)
	defer testutil.CloseDatabase()

	mockServer := testutil.CreateServer(http.StatusOK, notFinishedMatchResponse)
	defer mockServer.Close()

	startTime, _ := time.Parse("2006-01-02 15:04:05", time.Now().Add(-22*time.Hour).Format("2006-01-02 15:04:05"))
	match := createMatch(t, startTime)

	err := Reconcile()
	if err != nil {
		t.Errorf("Expected no error but got: %v", err)
	}

	actualMatch, err := repository.FindByCanonicalID(match.CanonicalID, entity.FootballOrg)
	if err != nil {
		t.Errorf("Expected no error but got: %v", err)
		return
	}

	if actualMatch == nil {
		t.Errorf("Expected match to be found, but it is nil")
		return
	}

	if actualMatch.Status != entity.Pending {
		t.Errorf("Expected match to be in pending status, but it is %d", actualMatch.Status)
	}

	if actualMatch.HomeTeamScore != 0 {
		t.Errorf("Expected match to have home team score 0, but it is %d", actualMatch.HomeTeamScore)
	}

	if actualMatch.AwayTeamScore != 0 {
		t.Errorf("Expected match to have away team score 0, but it is %d", actualMatch.AwayTeamScore)
	}

	testutil.ExpectNumberOfRequests(t, mockServer, 1)
}

func Test_we_update_the_match_if_the_status_is_finished(t *testing.T) {
	testutil.InitDatabase(t)
	defer testutil.CloseDatabase()

	mockServer := testutil.CreateServer(http.StatusOK, finishedMatchResponse)
	defer mockServer.Close()

	startTime, _ := time.Parse("2006-01-02 15:04:05", time.Now().Add(-22*time.Hour).Format("2006-01-02 15:04:05"))
	match := createMatch(t, startTime)

	err := Reconcile()
	if err != nil {
		t.Errorf("Expected no error but got: %v", err)
	}

	actualMatch, err := repository.FindByCanonicalID(match.CanonicalID, entity.FootballOrg)
	if err != nil {
		t.Errorf("Expected no error but got: %v", err)
		return
	}

	if actualMatch == nil {
		t.Errorf("Expected match to be found, but it is nil")
		return
	}

	if actualMatch.Status != entity.Finished {
		t.Errorf("Expected match to be in finished status, but it is %d", actualMatch.Status)
	}

	if actualMatch.HomeTeamScore != 0 {
		t.Errorf("Expected match to have home team score 0, but it is %d", actualMatch.HomeTeamScore)
	}

	if actualMatch.AwayTeamScore != 3 {
		t.Errorf("Expected match to have away team score 3, but it is %d", actualMatch.AwayTeamScore)
	}

	testutil.ExpectNumberOfRequests(t, mockServer, 1)
}

func Test_we_continue_when_api_call_fails_during_reconciliation(t *testing.T) {
	testutil.InitDatabase(t)
	defer testutil.CloseDatabase()

	mockServer := testutil.CreateServer(http.StatusInternalServerError, "")
	defer mockServer.Close()

	startTime, _ := time.Parse("2006-01-02 15:04:05", time.Now().Add(-22*time.Hour).Format("2006-01-02 15:04:05"))
	match := createMatch(t, startTime)

	err := Reconcile()
	if err != nil {
		t.Errorf("Expected no error but got: %v", err)
	}

	// Match should still be in pending status since API call failed
	actualMatch, err := repository.FindByCanonicalID(match.CanonicalID, entity.FootballOrg)
	if err != nil {
		t.Errorf("Expected no error but got: %v", err)
		return
	}

	if actualMatch == nil {
		t.Errorf("Expected match to be found, but it is nil")
		return
	}

	if actualMatch.Status != entity.Pending {
		t.Errorf("Expected match to still be in pending status, but it is %d", actualMatch.Status)
	}

	testutil.ExpectNumberOfRequests(t, mockServer, 1)
}

func createMatch(t *testing.T, startTime time.Time) entity.Match {
	tx, _ := testutil.BeginTransaction(t)

	match := entity.NewMatch(
		startTime,
		entity.FootballOrg,
		"544391",
		entity.AthleticClub,
		entity.RealMadrid,
		0,
		0,
		entity.LaLiga,
	)
	repository.Save(tx, match)
	tx.Commit()

	return match
}
