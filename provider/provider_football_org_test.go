package main

import (
	"net/http"
	"strings"
	"testing"
)

func Test_we_can_handle_unauthorized_response(t *testing.T) {
	logger := getLogger()
    mockServer := createServer(http.StatusForbidden)
    defer mockServer.Close()

	args := []string{"provider", "football_org"}

	exitCode := run(args)
	if exitCode != 1 {
		t.Errorf("Expected exit code 1, but got %d", exitCode)
	}

    outputStr := logger.String()
	if !strings.Contains(outputStr, "403 Forbidden") {
		t.Errorf("Expected '403 Forbidden' in output, but got: %s", outputStr)
	}
}

// func TestSyncMatches_ServerError(t *testing.T) {
// 	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		w.WriteHeader(http.StatusInternalServerError)
// 		w.Write([]byte("Internal Server Error"))
// 	}))
// 	defer mockServer.Close()

// 	client := &FootballOrgClient{
// 		Client:      &http.Client{},
// 		APIEndpoint: mockServer.URL,
// 		APIKey:      "test-api-key",
// 	}

// 	// Capture output
// 	originalStdout := os.Stdout
// 	r, w, _ := os.Pipe()
// 	os.Stdout = w

// 	client.SyncMatches()

// 	w.Close()
// 	os.Stdout = originalStdout

// 	output := make([]byte, 1024)
// 	n, _ := r.Read(output)
// 	outputStr := string(output[:n])

// 	// Should still log the response body even on error
// 	if !strings.Contains(outputStr, "Internal Server Error") {
// 		t.Error("Expected response body to be logged even on server error")
// 	}
// }

// // TestSyncMatches_EmptyResponse tests empty response body
// func TestSyncMatches_EmptyResponse(t *testing.T) {
// 	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		w.WriteHeader(http.StatusOK)
// 		w.Write([]byte("{}"))
// 	}))
// 	defer mockServer.Close()

// 	client := &FootballOrgClient{
// 		Client:      &http.Client{},
// 		APIEndpoint: mockServer.URL,
// 		APIKey:      "test-api-key",
// 	}

// 	// Capture output
// 	originalStdout := os.Stdout
// 	r, w, _ := os.Pipe()
// 	os.Stdout = w

// 	client.SyncMatches()

// 	w.Close()
// 	os.Stdout = originalStdout

// 	output := make([]byte, 1024)
// 	n, _ := r.Read(output)
// 	outputStr := string(output[:n])

// 	if !strings.Contains(outputStr, "{}") {
// 		t.Error("Expected empty JSON response to be logged")
// 	}
// }

// // TestSyncMatches_InvalidURL tests URL parsing error
// func TestSyncMatches_InvalidURL(t *testing.T) {
// 	client := &FootballOrgClient{
// 		Client:      &http.Client{},
// 		APIEndpoint: "://invalid-url", // Invalid URL format
// 		APIKey:      "test-api-key",
// 	}

// 	// Capture stderr
// 	originalStderr := os.Stderr
// 	r, w, _ := os.Pipe()
// 	os.Stdout = w

// 	client.SyncMatches()

// 	w.Close()
// 	os.Stdout = originalStderr

// 	output := make([]byte, 1024)
// 	n, _ := r.Read(output)
// 	outputStr := string(output[:n])

// 	if !strings.Contains(outputStr, "failed to parse base URL") {
// 		t.Error("Expected error message about URL parsing failure")
// 	}
// }

// func TestSyncMatches_Success(t *testing.T) {
// 	// Create a mock server that returns a successful response
// 	mockResponse := map[string]interface{}{
// 		"matches": []map[string]interface{}{
// 			{
// 				"id":       1,
// 				"status":   "FINISHED",
// 				"homeTeam": map[string]string{"name": "Team A"},
// 				"awayTeam": map[string]string{"name": "Team B"},
// 				"score":    map[string]interface{}{"fullTime": map[string]int{"home": 2, "away": 1}},
// 			},
// 		},
// 	}

// 	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		// Verify the request
// 		if r.Method != "GET" {
// 			t.Errorf("Expected GET request, got %s", r.Method)
// 		}

// 		if r.Header.Get("X-Auth-Token") == "" {
// 			t.Error("Expected X-Auth-Token header to be set")
// 		}

// 		// Check that dateFrom and dateTo are in the query parameters
// 		if r.URL.Query().Get("dateFrom") == "" {
// 			t.Error("Expected dateFrom query parameter")
// 		}
// 		if r.URL.Query().Get("dateTo") == "" {
// 			t.Error("Expected dateTo query parameter")
// 		}

// 		// Return successful response
// 		w.Header().Set("Content-Type", "application/json")
// 		w.WriteHeader(http.StatusOK)
// 		json.NewEncoder(w).Encode(mockResponse)
// 	}))
// 	defer mockServer.Close()

// 	// Create client with mocked server
// 	client := &FootballOrgClient{
// 		Client:      &http.Client{},
// 		APIEndpoint: mockServer.URL,
// 		APIKey:      "test-api-key",
// 	}

// 	// Capture output
// 	originalStdout := os.Stdout
// 	r, w, _ := os.Pipe()
// 	os.Stdout = w

// 	// Run the function
// 	client.SyncMatches()

// 	// Restore stdout
// 	w.Close()
// 	os.Stdout = originalStdout

// 	// Read captured output
// 	output := make([]byte, 1024)
// 	n, _ := r.Read(output)
// 	outputStr := string(output[:n])

// 	// Verify output contains expected information
// 	if !strings.Contains(outputStr, "[INFO] Sending GET request") {
// 		t.Error("Expected log message about sending GET request")
// 	}
// 	if !strings.Contains(outputStr, "matches") {
// 		t.Error("Expected response body to contain 'matches'")
// 	}
// }

// // TestSyncMatches_RealisticResponse tests with a realistic API response structure
// func TestSyncMatches_RealisticResponse(t *testing.T) {
// 	realisticResponse := `{
// 		"matches": [
// 			{
// 				"id": 12345,
// 				"utcDate": "2024-01-15T20:00:00Z",
// 				"status": "FINISHED",
// 				"matchday": 1,
// 				"stage": "REGULAR_SEASON",
// 				"group": null,
// 				"lastUpdated": "2024-01-15T22:30:00Z",
// 				"homeTeam": {
// 					"id": 81,
// 					"name": "FC Barcelona"
// 				},
// 				"awayTeam": {
// 					"id": 86,
// 					"name": "Real Madrid"
// 				},
// 				"score": {
// 					"winner": "HOME_TEAM",
// 					"duration": "REGULAR",
// 					"fullTime": {
// 						"home": 3,
// 						"away": 1
// 					},
// 					"halfTime": {
// 						"home": 2,
// 						"away": 0
// 					}
// 				}
// 			}
// 		],
// 		"resultSet": {
// 			"count": 1,
// 			"first": "2024-01-15",
// 			"last": "2024-01-15",
// 			"played": 1
// 		}
// 	}`

// 	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		w.Header().Set("Content-Type", "application/json")
// 		w.WriteHeader(http.StatusOK)
// 		w.Write([]byte(realisticResponse))
// 	}))
// 	defer mockServer.Close()

// 	client := &FootballOrgClient{
// 		Client:      &http.Client{},
// 		APIEndpoint: mockServer.URL,
// 		APIKey:      "test-api-key",
// 	}

// 	// Capture output
// 	originalStdout := os.Stdout
// 	r, w, _ := os.Pipe()
// 	os.Stdout = w

// 	client.SyncMatches()

// 	w.Close()
// 	os.Stdout = originalStdout

// 	output := make([]byte, 4096)
// 	n, _ := r.Read(output)
// 	outputStr := string(output[:n])

// 	// Verify realistic response is logged
// 	if !strings.Contains(outputStr, "FC Barcelona") {
// 		t.Error("Expected response to contain team name from realistic response")
// 	}
// 	if !strings.Contains(outputStr, "Real Madrid") {
// 		t.Error("Expected response to contain away team name")
// 	}
// }
