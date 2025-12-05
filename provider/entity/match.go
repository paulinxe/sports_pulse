package entity

import (
    "log/slog"
    "time"
)

type Match struct {
    ID              string
    Start           time.Time
    End             time.Time
    Status          string
    Provider        Provider
    ProviderMatchID string
    HomeTeamID      int
    AwayTeamID      int
    HomeTeamScore   int
    AwayTeamScore   int
}

func NewMatch(
    start string,
    end string,
    provider Provider,
    providerMatchID string,
    homeTeamID int,
    awayTeamID int,
    homeTeamScore int,
    awayTeamScore int,
) Match {
    id := "1" // TODO: we need to use our Canonical way of generating the id.

    startTime, err := time.Parse(time.RFC3339, start)
    if err != nil {
        // TODO: if we can't parse the date, we should log an error. we will manually need to set the start time of the match.
        slog.Warn("Failed to parse match date, using current time",
            "match_id", id,
            "start", start,
            "error", err)
            startTime = time.Now() // TODO: we need to set the start time of the match.
    }

    // TODO: avoid magic numbers
    endTime := startTime.Add(2 * time.Hour)

    return Match{
        ID: id,
        Start: startTime,
        End: endTime,
        Status: "pending",
        Provider: provider,
        ProviderMatchID: providerMatchID,
        HomeTeamID: homeTeamID,
        AwayTeamID: awayTeamID,
        HomeTeamScore: homeTeamScore,
        AwayTeamScore: awayTeamScore,
    }
}