package football_org

import (
    "encoding/json"
    "fmt"
    "io"
    "log/slog"
    "net/http"
    "net/url"
    "os"
    "provider/entity"
    "provider/repository"
    "time"
)

func Sync() error {
	body, err := fetchAPI()
	if err != nil {
		return err
	}

    var matchesResponse MatchesResponse
    if err := json.Unmarshal(body, &matchesResponse); err != nil {
        slog.Error("Failed to parse JSON response", "error", err, "body", string(body))
        return fmt.Errorf("failed to parse JSON response: %v", err)
    }

    slog.Debug(fmt.Sprintf("Successfully parsed %d matches", len(matchesResponse.Matches)))

    for _, footballOrgMatch := range matchesResponse.Matches {
        homeTeamID, ok := FootballOrgTeamMapping[footballOrgMatch.HomeTeam.ID]
        if !ok {
            slog.Error("Failed to map home team ID, skipping match",
                "external_team_id", footballOrgMatch.HomeTeam.ID,
                "match_id", footballOrgMatch.ID)
            continue
        }
        awayTeamID, ok := FootballOrgTeamMapping[footballOrgMatch.AwayTeam.ID]
        if !ok {
            slog.Error("Failed to map away team ID, skipping match",
                "external_team_id", footballOrgMatch.AwayTeam.ID,
                "match_id", footballOrgMatch.ID)
            continue
        }

        match := entity.NewMatch(
            footballOrgMatch.UTCDate,
            entity.FootballOrg,
            fmt.Sprintf("%d", footballOrgMatch.ID),
            homeTeamID,
            awayTeamID,
            footballOrgMatch.Score.FullTime.Home,
            footballOrgMatch.Score.FullTime.Away,
            entity.LaLiga, // TODO: we need to map this value
        )

        if err := repository.Save(match); err != nil {
            return fmt.Errorf("failed to insert matches: %v", err)
        }
    }

    slog.Debug(fmt.Sprintf("Successfully inserted %d matches into database", len(matchesResponse.Matches)))
    return nil
}

func fetchAPI() ([]byte, error) {
	apiEndpoint := os.Getenv("FOOTBALL_ORG_API_ENDPOINT")
    apiKey := os.Getenv("FOOTBALL_ORG_API_KEY")

    // LaLiga
    base, err := url.Parse(apiEndpoint + "/competitions/2014/matches")
    if err != nil {
        slog.Error("Failed to parse base URL", "error", err)
        return nil, fmt.Errorf("failed to parse base URL: %v", err)
    }

    // TODO: calculations (make sure this algorithm is correct)
    // from: here we should go to the db to get the most recent match stored.
    //  If we don't have any, we start from now.
    //  If from is already one week from now, we stop execution.
    // to: we add 1 week to from.
    from := time.Now()

    // Calculate the time 7 days (1 week) from now
    to := from.Add(7 * 24 * time.Hour)

    params := url.Values{}
    params.Add("dateFrom", from.Format("2006-01-02"))
    params.Add("dateTo", to.Format("2006-01-02"))
    //params.Add("status", "FINISHED")

    // Encode the parameters and append them to the base URL
    base.RawQuery = params.Encode()
    finalURL := base.String()

    slog.Debug("Sending GET request", "url", finalURL)

    // Create a new HTTP request with custom headers
    req, err := http.NewRequest("GET", finalURL, nil)
    if err != nil {
        slog.Error("Failed to create request", "error", err)
        return nil, fmt.Errorf("failed to create request: %v", err)
    }

    req.Header.Set("X-Auth-Token", apiKey)

    // Execute the GET request
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

    // Check for HTTP error status codes
    if resp.StatusCode >= 400 {
        slog.Error("HTTP error response",
            "status_code", resp.StatusCode,
            "status", resp.Status,
            "body", string(body))
        return nil, fmt.Errorf("HTTP error: %d %s", resp.StatusCode, resp.Status)
    }

    //slog.Info("Response received", "body", string(body))
	return body, nil
}