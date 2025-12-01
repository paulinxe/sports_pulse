package main

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"time"
)

// HTTPClient interface for dependency injection in tests
// TODO: we need to move this to somewhere else
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// FootballOrgClient wraps the HTTP client and API configuration
type FootballOrgClient struct {
	Client      HTTPClient
	APIEndpoint string
	APIKey      string
}

// NewFootballOrgClient creates a new client with default HTTP client
// Uses FOOTBALL_ORG_API_ENDPOINT env var if set, otherwise defaults to the production API
func NewFootballOrgClient() *FootballOrgClient {
	return &FootballOrgClient{
		Client:      &http.Client{},
		APIEndpoint: os.Getenv("FOOTBALL_ORG_API_ENDPOINT"),
		APIKey:      os.Getenv("FOOTBALL_ORG_API_KEY"),
	}
}

// sync_football_org fetches matches using the client configured via environment variables
// Uses FOOTBALL_ORG_API_ENDPOINT env var if set, otherwise defaults to production API
func sync_football_org() error {
	client := NewFootballOrgClient()
	return client.SyncMatches()
}

// SyncMatches fetches matches from the Football Data API
// TODO: we need a better error handling everywhere
func (c *FootballOrgClient) SyncMatches() error {
	// LaLiga
	base, err := url.Parse(c.APIEndpoint + "/competitions/2014/matches")
	if err != nil {
		return buildError(fmt.Sprintf("failed to parse base URL: %v", err))
	}

	// TODO: calculations (make sure this algorithm is correct)
	// from: here we should go to the db to get the most recent match stored.
	//  If we don't have any, we start from now.
	//  If from is already one week from now, we stop execution.
	// to: we add 1 week to from.
	from := time.Now()

	// Calculate the time 7 days (1 week) from now
	to := from.Add(7 * 24 * time.Hour)

	// 2. Create a new Values object for query parameters.
	// This is the cleanest way to handle query parameters as it automatically handles URL encoding.
	params := url.Values{}
	// Add your query parameters here
	params.Add("dateFrom", from.Format("2006-01-02"))
	params.Add("dateTo", to.Format("2006-01-02"))
	//params.Add("status", "FINISHED")

	// 3. Encode the parameters and append them to the base URL.
	base.RawQuery = params.Encode()
	finalURL := base.String()

	slog.Info("Sending GET request", "url", finalURL)

	// 4. Create a new HTTP request with custom headers
	req, err := http.NewRequest("GET", finalURL, nil)
	if err != nil {
		return buildError(fmt.Sprintf("failed to create request: %v", err))
	}

	// Add the X-Auth-Token header
	req.Header.Set("X-Auth-Token", c.APIKey)

	// 5. Execute the GET request
	resp, err := c.Client.Do(req)
	if err != nil {
		return buildError(fmt.Sprintf("failed to get matches: %v", err))
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return buildError(fmt.Sprintf("failed to read response body: %v", err))
	}

	// Check for HTTP error status codes
	if resp.StatusCode >= 400 {
		slog.Error("HTTP error response",
			"status_code", resp.StatusCode,
			"status", resp.Status,
			"body", string(body))
		return buildError(fmt.Sprintf("HTTP error: %d %s", resp.StatusCode, resp.Status))
	}

	slog.Info("Response received", "body", string(body))
	return nil
}
