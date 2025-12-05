package football_org

import (
    _ "embed"
    "net/http"
    "provider/db"
    "provider/testutil"
    "strings"
    "testing"
)

//go:embed sample.json
var successResponse string

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

func Test_we_can_handle_valid_response(t *testing.T) {
    logger := testutil.GetLogger()
    mockServer := testutil.CreateServer(http.StatusOK, successResponse)
    defer mockServer.Close()

    // Initialize database - fail test if DB can't be initialized
    // TODO: move this somewhere else.
    if err := testutil.InitDatabase(); err != nil {
        t.Fatalf("Failed to initialize database: %v", err)
    }
    defer testutil.CloseDatabase()

    // Verify database connection is ready
    if db.DB == nil {
        t.Fatal("Database connection is nil after initialization")
    }

    // Clean up before the test
    // TODO: we need a better way to clean up the database.
    _, _ = db.DB.Exec("DELETE FROM matches")

    err := Sync()
    if err != nil {
        outputStr := logger.String()
        t.Logf("=== Full Log Output ===")
        t.Logf("%s", outputStr)
        t.Logf("=======================")
        t.Errorf("Expected no error but got: %v", err)
        return
    }

    outputStr := logger.String()

    if !strings.Contains(outputStr, "Successfully parsed 1 matches") {
        t.Errorf("Expected 'Successfully parsed 1 matches' in output, but got: %s", outputStr)
    }

    // Verify the record was created in the database
    var count int
    err = db.DB.QueryRow("SELECT COUNT(*) FROM matches WHERE id = $1", "1").Scan(&count)
    if err != nil {
        t.Fatalf("Failed to query database: %v", err)
    }

    if count != 1 {
        t.Errorf("Expected 1 record with id '544391', but found %d", count)
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
