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

//go:embed test_data_provider/invalid_match_date.json
var invalidMatchDateResponse string

func Test_we_can_handle_unauthorized_response(t *testing.T) {
    // TODO: find a way to not duplicate the database initialization and cleanup code.
    testutil.InitDatabase(t)
    defer testutil.CloseDatabase()

    logger := testutil.GetLogger()
    mockServer := testutil.CreateServer(http.StatusForbidden, "")
    defer mockServer.Close()

    err := Sync(entity.LaLiga)
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

    err := Sync(entity.LaLiga)
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

    err := Sync(entity.LaLiga)
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

    err := Sync(entity.LaLiga)
    if err == nil {
        t.Error("Expected error but got nil")
    }

    outputStr := logger.String()
    if !strings.Contains(outputStr, "Failed to parse JSON response") {
        t.Errorf("Expected 'Failed to parse JSON response' in output, but got: %s", outputStr)
    }
}

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
    mockServer := testutil.CreateServer(http.StatusOK, homeTeamNotMappedResponse)
    defer mockServer.Close()

    err := Sync(entity.LaLiga)
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

    err := Sync(entity.LaLiga)
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

    err := Sync(entity.LaLiga)
    if err != nil {
        t.Errorf("Expected no error but got: %v", err)
    }

    // When no matches exist, dateFrom should be today and dateTo should be 3 days from today
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

    canonicalID := "58a49d03246d65ce3ce64dd7ca690977fe0f2feeccf3403ebe8b95e515599ff8"
    actualMatch, err := repository.FindByCanonicalID(canonicalID, entity.FootballOrg)
    if err != nil {
        t.Errorf("Expected no error but got: %v", err)
        return
    }

    if actualMatch == nil {
        t.Errorf("Expected match to be found, but it is nil")
        return
    }

    expectedMatch := entity.Match{
        ID:              actualMatch.ID, // Small hack to be able to compare the matches
        CanonicalID:     canonicalID,
        Start:           expectedMatchStart,
        End:             expectedMatchEnd,
        Status:          "pending",
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

func Test_no_api_call_is_made_when_last_match_is_already_3_days_in_the_future(t *testing.T) {
    testutil.InitDatabase(t)
    defer testutil.CloseDatabase()

    mockServer := testutil.CreateServer(http.StatusOK, "")
    defer mockServer.Close()

    futureDate := time.Now().Add(3 * 24 * time.Hour).Add(1 * time.Minute)
    match := entity.NewMatch(
        futureDate,
        entity.FootballOrg,
        "1",
        entity.AthleticClub,
        entity.RealMadrid,
        0,
        0,
        entity.LaLiga,
    )

    tx, _ := testutil.BeginTransaction(t)
    repository.Save(tx, match)
    tx.Commit()
    
    err := Sync(entity.LaLiga)
    if err != nil {
        t.Errorf("Expected no error but got: %v", err)
    }

    testutil.ExpectNumberOfRequests(t, mockServer, 0)
}

func Test_we_can_handle_invalid_match_date(t *testing.T) {
    testutil.InitDatabase(t)
    defer testutil.CloseDatabase()

    logger := testutil.GetLogger()
    mockServer := testutil.CreateServer(http.StatusOK, invalidMatchDateResponse)
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

// TODO: add test for when matches exist for the competition