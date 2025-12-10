package sync

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"provider/football_org/api"
	"time"
)

const DAYS_TO_FETCH = 3

func FetchAPI(competitionID int, mostRecentTimestamp *time.Time) (api.MatchesResponse, error) {
	apiURL, err := buildAPIURL(competitionID, mostRecentTimestamp)
	if err != nil {
		return api.MatchesResponse{}, err
	}

	body, err := makeHTTPRequest(apiURL)
	if err != nil {
		return api.MatchesResponse{}, err
	}

	return parseMatchesResponse(body)
}

func buildAPIURL(competitionID int, mostRecentTimestamp *time.Time) (string, error) {
	apiEndpoint := os.Getenv("FOOTBALL_ORG_API_ENDPOINT")
	baseURL := apiEndpoint + fmt.Sprintf("/competitions/%d/matches", competitionID)

	base, err := url.Parse(baseURL)
	if err != nil {
		slog.Error("Failed to parse base URL", "error", err)
		return "", fmt.Errorf("failed to parse base URL: %v", err)
	}

	from, to := calculateDateRange(mostRecentTimestamp)
	params := url.Values{}
	params.Add("dateFrom", from.Format("2006-01-02"))
	params.Add("dateTo", to.Format("2006-01-02"))

	base.RawQuery = params.Encode()
	return base.String(), nil
}

func calculateDateRange(mostRecentTimestamp *time.Time) (time.Time, time.Time) {
	var from time.Time
	if mostRecentTimestamp == nil {
		from = time.Now()
	} else {
		from = mostRecentTimestamp.Add(24 * time.Hour)
	}

	to := from.Add(DAYS_TO_FETCH * 24 * time.Hour)
	return from, to
}

func makeHTTPRequest(url string) ([]byte, error) {
	apiKey := os.Getenv("FOOTBALL_ORG_API_KEY")

	slog.Debug("Sending GET request", "url", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		slog.Error("Failed to create request", "error", err)
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("X-Auth-Token", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("Failed to get matches", "error", err)
		return nil, fmt.Errorf("failed to get matches: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("Failed to read response body", "error", err)
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	if resp.StatusCode >= 400 {
		slog.Error("HTTP error response",
			"status_code", resp.StatusCode,
			"status", resp.Status,
			"body", string(body))
		return nil, fmt.Errorf("HTTP error: %d %s", resp.StatusCode, resp.Status)
	}

	return body, nil
}

func parseMatchesResponse(body []byte) (api.MatchesResponse, error) {
	var matchesResponse api.MatchesResponse
	if err := json.Unmarshal(body, &matchesResponse); err != nil {
		slog.Error("Failed to parse JSON response", "error", err, "body", string(body))
		return api.MatchesResponse{}, fmt.Errorf("failed to parse JSON response: %v", err)
	}

	slog.Debug(fmt.Sprintf("Successfully parsed %d matches", len(matchesResponse.Matches)))
	return matchesResponse, nil
}
