package football_org

import (
    "encoding/json"
    "fmt"
    "io"
    "log/slog"
    "net/http"
    "net/url"
    "os"
    "provider/db"
    "time"
)

func Sync() error {
    apiEndpoint := os.Getenv("FOOTBALL_ORG_API_ENDPOINT")
    apiKey := os.Getenv("FOOTBALL_ORG_API_KEY")

    // LaLiga
    base, err := url.Parse(apiEndpoint + "/competitions/2014/matches")
    if err != nil {
        slog.Error("Failed to parse base URL", "error", err)
        return fmt.Errorf("failed to parse base URL: %v", err)
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
        return fmt.Errorf("failed to create request: %v", err)
    }

    req.Header.Set("X-Auth-Token", apiKey)

    // Execute the GET request
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        slog.Error("Failed to get matches", "error", err)
        return fmt.Errorf("failed to get matches: %v", err)
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        slog.Error("Failed to read response body", "error", err)
        return fmt.Errorf("failed to read response body: %v", err)
    }

    // Check for HTTP error status codes
    if resp.StatusCode >= 400 {
        slog.Error("HTTP error response",
            "status_code", resp.StatusCode,
            "status", resp.Status,
            "body", string(body))
        return fmt.Errorf("HTTP error: %d %s", resp.StatusCode, resp.Status)
    }

    //slog.Info("Response received", "body", string(body))

    var matchesResponse MatchesResponse
    if err := json.Unmarshal(body, &matchesResponse); err != nil {
        slog.Error("Failed to parse JSON response", "error", err, "body", string(body))
        return fmt.Errorf("failed to parse JSON response: %v", err)
    }

    slog.Info(fmt.Sprintf("Successfully parsed %d matches", len(matchesResponse.Matches)))

    // Insert matches into database
    if err := insertMatches(matchesResponse.Matches); err != nil {
        return fmt.Errorf("failed to insert matches: %v", err)
    }

    slog.Info(fmt.Sprintf("Successfully inserted %d matches into database", len(matchesResponse.Matches)))
    return nil
}

func insertMatches(matches []Match) error {
    if db.DB == nil {
        slog.Warn("Database connection not initialized, skipping insert")
        return nil // Skip database operations if DB is not available (e.g., in tests)
    }

    // TODO: use a db transaction
    // TODO: see if we can use named parameters
    query := `
        INSERT INTO matches (
            id, home_team_id, away_team_id, start, "end", status,
            home_team_score, away_team_score, provider_match_id,
            provider, transaction_hash, signature
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
    `

    for _, match := range matches {
        startTime, err := time.Parse(time.RFC3339, match.UTCDate)
        if err != nil {
            // TODO: if we can't parse the date, we should log an error.
            // TODO: we manually need to set the start time of the match.
            slog.Warn("Failed to parse match date, using current time",
                "match_id", match.ID,
                "utc_date", match.UTCDate,
                "error", err)
            startTime = time.Now()
        }

        // Calculate end time (start + 2 hours as dummy value)
        endTime := startTime.Add(2 * time.Hour)

        // Map status - use match.Status directly, assuming it matches the enum
        // TODO: we need a mapping of the statuses from football_org to our enum.
        status := "pending"

        homeScore := int64(match.Score.FullTime.Home)
        awayScore := int64(match.Score.FullTime.Away)

        // Use provider_match_id as the primary key id (unique per match)
        // TODO: we need to use our Canonical way of generating the id.
        providerMatchID := fmt.Sprintf("%d", match.ID)

        // Execute insert
        result, err := db.DB.Exec(query,
            providerMatchID,   // id (use provider_match_id for uniqueness)
            match.HomeTeam.ID, // home_team_id
            match.AwayTeam.ID, // away_team_id
            startTime,         // start
            endTime,           // end
            status,            // status
            homeScore,         // home_team_score
            awayScore,         // away_team_score
            providerMatchID,   // provider_match_id (convert int to string)
            1,                 // provider (dummy INT)
            "test",            // transaction_hash (dummy VARCHAR)
            "test",            // signature (dummy VARCHAR)
        )

        if err != nil {
            slog.Error("Failed to insert match",
                "match_id", match.ID,
                "error", err)
            return fmt.Errorf("failed to insert match %d: %v", match.ID, err)
        }

        rowsAffected, _ := result.RowsAffected()
        slog.Debug("Inserted match",
            "match_id", match.ID,
            "rows_affected", rowsAffected)

        slog.Debug("Inserted match",
            "match_id", match.ID,
            "home_team_id", match.HomeTeam.ID,
            "away_team_id", match.AwayTeam.ID)
    }

    return nil
}
