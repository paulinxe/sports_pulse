package football_org

import (
    _ "embed"
    "net/http"
    "provider/testutil"
    "strings"
    "testing"
)

//go:embed test_data_provider/valid_response.json
var successResponse string
//go:embed test_data_provider/home_team_not_mapped.json
var homeTeamNotMappedResponse string
//go:embed test_data_provider/away_team_not_mapped.json
var awayTeamNotMappedResponse string

func Test_we_can_handle_unauthorized_response(t *testing.T) {
    logger := testutil.GetLogger()
    mockServer := testutil.CreateServer(http.StatusForbidden, "")
    defer mockServer.Close()

    err := Sync()
    if err == nil {
        t.Error("Expected error but got nil")
    }

    outputStr := logger.String()
    if !strings.Contains(outputStr, "403 Forbidden") {
        t.Errorf("Expected '403 Forbidden' in output, but got: %s", outputStr)
    }
}

func Test_we_can_handle_too_many_requests_response(t *testing.T) {
    logger := testutil.GetLogger()
    mockServer := testutil.CreateServer(http.StatusTooManyRequests, "")
    defer mockServer.Close()

    err := Sync()
    if err == nil {
        t.Error("Expected error but got nil")
    }

    outputStr := logger.String()
    if !strings.Contains(outputStr, "429 Too Many Requests") {
        t.Errorf("Expected '429 Too Many Requests' in output, but got: %s", outputStr)
    }
}

func Test_we_can_handle_internal_server_error_response(t *testing.T) {
    logger := testutil.GetLogger()
    mockServer := testutil.CreateServer(http.StatusInternalServerError, "")
    defer mockServer.Close()

    err := Sync()
    if err == nil {
        t.Error("Expected error but got nil")
    }

    outputStr := logger.String()
    if !strings.Contains(outputStr, "500 Internal Server Error") {
        t.Errorf("Expected '500 Internal Server Error' in output, but got: %s", outputStr)
    }
}

func Test_we_skip_the_match_if_home_team_is_not_mapped(t *testing.T) {
	logger := testutil.GetLogger()
    mockServer := testutil.CreateServer(http.StatusOK, homeTeamNotMappedResponse)
    defer mockServer.Close()

    testutil.InitDatabase(t)
    defer testutil.CloseDatabase()

    err := Sync()
    if err != nil {
        t.Errorf("Expected no error but got: %v", err)
    }

    outputStr := logger.String()

	if !strings.Contains(outputStr, "Failed to map home team ID, skipping match") {
		t.Errorf("Expected 'Failed to map home team ID, skipping match' in output, but got: %s", outputStr)
	}

    // We should still create the Athletic - Real Madrid match
    if !testutil.MatchExists(t, "58a49d03246d65ce3ce64dd7ca690977fe0f2feeccf3403ebe8b95e515599ff8") {
        t.Errorf("Athletic - Real Madrid match should exist, but it does not")
    }
}

func Test_we_skip_the_match_if_away_team_is_not_mapped(t *testing.T) {
	logger := testutil.GetLogger()
    mockServer := testutil.CreateServer(http.StatusOK, awayTeamNotMappedResponse)
    defer mockServer.Close()

    testutil.InitDatabase(t)
    defer testutil.CloseDatabase()

	err := Sync()
    if err != nil {
        t.Errorf("Expected no error but got: %v", err)
    }

    outputStr := logger.String()
    if !strings.Contains(outputStr, "Failed to map away team ID, skipping match") {
        t.Errorf("Expected 'Failed to map away team ID, skipping match' in output, but got: %s", outputStr)
    }

    // We should still create the Athletic - Real Madrid match
    if !testutil.MatchExists(t, "58a49d03246d65ce3ce64dd7ca690977fe0f2feeccf3403ebe8b95e515599ff8") {
        t.Errorf("Athletic - Real Madrid match should exist, but it does not")
    }
}

func Test_we_can_handle_valid_response(t *testing.T) {
    logger := testutil.GetLogger()
    mockServer := testutil.CreateServer(http.StatusOK, successResponse)
    defer mockServer.Close()

    testutil.InitDatabase(t)
    defer testutil.CloseDatabase()

    err := Sync()
    if err != nil {
        t.Errorf("Expected no error but got: %v", err)
    }

    outputStr := logger.String()

    if !strings.Contains(outputStr, "Successfully parsed 1 matches") {
        t.Errorf("Expected 'Successfully parsed 1 matches' in output, but got: %s", outputStr)
    }

    // Verify the record was created in the database
    if !testutil.MatchExists(t, "58a49d03246d65ce3ce64dd7ca690977fe0f2feeccf3403ebe8b95e515599ff8") {
        t.Errorf("Athletic - Real Madrid match should exist, but it does not")
    }

    // Optionally, verify some fields of the created record
    // var homeTeamID, awayTeamID int
    // var status string
    // err = db.DB.QueryRow(
    // 	"SELECT home_team_id, away_team_id, status FROM matches WHERE id = $1",
    // 	"544391",
    // ).Scan(&homeTeamID, &awayTeamID, &status)
    // if err != nil {
    // 	if err == sql.ErrNoRows {
    // 		t.Error("Record with id '544391' was not found in database")
    // 	} else {
    // 		t.Fatalf("Failed to query record details: %v", err)
    // 	}
    // 	return
    // }

    // // Verify the record has expected values from sample.json
    // // From sample.json: homeTeam.id=77, awayTeam.id=86
    // if homeTeamID != 77 {
    // 	t.Errorf("Expected home_team_id to be 77, but got %d", homeTeamID)
    // }
    // if awayTeamID != 86 {
    // 	t.Errorf("Expected away_team_id to be 86, but got %d", awayTeamID)
    // }
}
