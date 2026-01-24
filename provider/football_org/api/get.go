package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"net/url"
	"time"
)

func GetList(ctx context.Context, competitionID uint, from time.Time, to time.Time) (MatchesResponse, error) {
	url := buildAPIPath(competitionID, from, to)
	body, err := get(ctx, url)
	if err != nil {
		return MatchesResponse{}, err
	}

	var match MatchesResponse
	if err := parseResponse(body, &match); err != nil {
		return MatchesResponse{}, err
	}

	return match, nil
}

func buildAPIPath(competitionID uint, from time.Time, to time.Time) string {
	path := fmt.Sprintf("/competitions/%d/matches", competitionID)

	params := url.Values{}
	params.Add("dateFrom", from.Format("2006-01-02"))
	params.Add("dateTo", to.Format("2006-01-02"))

	queryString := params.Encode()
	return path + "?" + queryString
}

func get(ctx context.Context, url string) ([]byte, error) {
	url = os.Getenv("FOOTBALL_ORG_API_ENDPOINT") + url
	apiKey := os.Getenv("FOOTBALL_ORG_API_KEY")

	slog.Debug("Sending GET request", "url", url)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-Auth-Token", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		if ctx.Err() == context.Canceled {
			return nil, fmt.Errorf("request canceled: %w", err)
		}

		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("context timeout: %w", err)
		}

		return nil, fmt.Errorf("failed to get matches: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP error: %d %s", resp.StatusCode, resp.Status)
	}

	return body, nil
}

func parseResponse(body []byte, parseTo interface{}) error {
	if err := json.Unmarshal(body, parseTo); err != nil {
		return fmt.Errorf("failed to parse JSON response: %w", err)
	}

	return nil
}