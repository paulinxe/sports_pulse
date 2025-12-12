package api

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"io"
	"encoding/json"
)

func GetList(url string) (MatchesResponse, error) {
	body, err := get(url)
	if err != nil {
		return MatchesResponse{}, err
	}

	var match MatchesResponse
	if err := parseResponse(body, &match); err != nil {
		return MatchesResponse{}, err
	}

	return match, nil
}

func GetOne(url string) (FootballOrgMatch, error) {
	body, err := get(url)
	if err != nil {
		return FootballOrgMatch{}, err
	}

	var match FootballOrgMatch
	if err := parseResponse(body, &match); err != nil {
		return FootballOrgMatch{}, err
	}

	return match, nil
}

func get(url string) ([]byte, error) {
	url = os.Getenv("FOOTBALL_ORG_API_ENDPOINT") + url
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

func parseResponse(body []byte, parseTo interface{}) error {
	if err := json.Unmarshal(body, parseTo); err != nil {
		slog.Error("Failed to parse JSON response", "error", err, "body", string(body))
		return fmt.Errorf("failed to parse JSON response: %v", err)
	}

	return nil
}