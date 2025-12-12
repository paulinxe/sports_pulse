package football_org

import (
	_ "embed"
	"net/http"	
	"provider/testutil"
	"testing"
	"time"
	"provider/entity"
	"provider/repository"
)

//go:embed test_data_provider/matches/not_finished_match.json
var notFinishedMatchResponse string

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

func Test_we_dont_update_the_match_if_the_status_is_not_finished(t *testing.T) {
	testutil.InitDatabase(t)
	defer testutil.CloseDatabase()

	mockServer := testutil.CreateServer(http.StatusOK, notFinishedMatchResponse)
	defer mockServer.Close()

	startTime, _ := time.Parse("2006-01-02 15:04:05", time.Now().Add(-22 * time.Hour).Format("2006-01-02 15:04:05"))
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
	tx, _ := testutil.BeginTransaction(t)
	repository.Save(tx, match)
	tx.Commit()

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