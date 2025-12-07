package football_org

import (
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

//go:embed test_data_provider/valid_response.json
var successResponse string

//go:embed test_data_provider/home_team_not_mapped.json
var homeTeamNotMappedResponse string

//go:embed test_data_provider/away_team_not_mapped.json
var awayTeamNotMappedResponse string

func Test_we_can_handle_unauthorized_response(t *testing.T) {
    // TODO: find a way to not duplicate the database initialization and cleanup code.
    testutil.InitDatabase(t)
    defer testutil.CloseDatabase()

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
    testutil.InitDatabase(t)
    defer testutil.CloseDatabase()

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
    testutil.InitDatabase(t)
    defer testutil.CloseDatabase()

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

func Test_we_can_handle_invalid_json_response(t *testing.T) {
    testutil.InitDatabase(t)
    defer testutil.CloseDatabase()

    logger := testutil.GetLogger()
    mockServer := testutil.CreateServer(http.StatusOK, "invalid json")
    defer mockServer.Close()

    err := Sync()
    if err == nil {
        t.Error("Expected error but got nil")
    }

    outputStr := logger.String()
    if !strings.Contains(outputStr, "Failed to parse JSON response") {
        t.Errorf("Expected 'Failed to parse JSON response' in output, but got: %s", outputStr)
    }
}

func Test_we_skip_the_match_if_home_team_is_not_mapped(t *testing.T) {
    testutil.InitDatabase(t)
    defer testutil.CloseDatabase()

    logger := testutil.GetLogger()
    mockServer := testutil.CreateServer(http.StatusOK, homeTeamNotMappedResponse)
    defer mockServer.Close()

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
    testutil.InitDatabase(t)
    defer testutil.CloseDatabase()

    logger := testutil.GetLogger()
    mockServer := testutil.CreateServer(http.StatusOK, awayTeamNotMappedResponse)
    defer mockServer.Close()

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

func Test_we_insert_a_match_when_no_matches_exist_for_competition(t *testing.T) {
    testutil.InitDatabase(t)
    defer testutil.CloseDatabase()

    mockServer := testutil.CreateServer(http.StatusOK, successResponse)
    defer mockServer.Close()

    err := Sync()
    if err != nil {
        t.Errorf("Expected no error but got: %v", err)
    }

    // TODO: we need to assert we asked the api with the correct date range

    start, _ := time.Parse("2006-01-02 15:04:05", "2025-12-03 18:00:00")
    end, _ := time.Parse("2006-01-02 15:04:05", "2025-12-03 20:00:00")

    expectedMatch := entity.Match{
        ID:              "58a49d03246d65ce3ce64dd7ca690977fe0f2feeccf3403ebe8b95e515599ff8",
        Start:           start,
        End:             end,
        Status:          "pending",
        Provider:        entity.FootballOrg,
        ProviderMatchID: "544391",
        CompetitionID:   entity.LaLiga,
        HomeTeamID:      entity.AthleticClub,
        AwayTeamID:      entity.RealMadrid,
        HomeTeamScore:   0,
        AwayTeamScore:   0,
    }
    actualMatch, err := repository.FindById("58a49d03246d65ce3ce64dd7ca690977fe0f2feeccf3403ebe8b95e515599ff8")
    if err != nil {
        t.Errorf("Expected no error but got: %v", err)
        return
    }

    if actualMatch == nil {
        t.Errorf("Expected match to be found, but it is nil")
        return
    }

    if !reflect.DeepEqual(*actualMatch, expectedMatch) {
        t.Errorf("Expected match %+v, but got %+v", expectedMatch, actualMatch)
    }
}
