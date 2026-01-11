package football_org

import (
	"context"
	"fmt"
	"log/slog"
	"provider/entity"
	"provider/football_org/sync"
	"provider/repository"
	"time"
)

const SYNC_CONTEXT_TIMEOUT = 15 * time.Second

// This Sync function is intended to run once a day to fetch the latest matches and insert them into the database.
// It will skip the sync if the most recent match is already 3+ days in the future.
// In reconcile.go we take care of updating the matches that are already in the database.
func Sync(competition entity.Competition) error {
	if err := validateCompetition(competition); err != nil {
		return err
	}

	mostRecentTimestamp, err := repository.FindMostRecentTimestamp(competition, entity.FootballOrg)
	if err != nil {
		return fmt.Errorf("failed to find most recent timestamp: %w", err)
	}

	if shouldSkipSync(mostRecentTimestamp) {
		return nil
	}

	// Create context with 15-second timeout for HTTP requests
	ctx, cancel := context.WithTimeout(context.Background(), SYNC_CONTEXT_TIMEOUT)
	defer cancel()

	competitionID := CompetitionToFootballOrgID[competition]
	matchesResponse, err := sync.FetchAPI(ctx, competitionID, mostRecentTimestamp)
	if err != nil {
		return err
	}

	if err := sync.SaveMatches(matchesResponse.Matches, competition, FootballOrgTeamMapping); err != nil {
		return err
	}

	slog.Debug(fmt.Sprintf("Successfully inserted %d matches into database", len(matchesResponse.Matches)))
	return nil
}

func validateCompetition(competition entity.Competition) error {
	if _, ok := CompetitionToFootballOrgID[competition]; !ok {
		return fmt.Errorf("unknown competition: %d", competition)
	}

	return nil
}

func shouldSkipSync(mostRecentTimestamp *time.Time) bool {
	if mostRecentTimestamp == nil {
		return false
	}

	now := time.Now()
	if mostRecentTimestamp.After(now.Add(3 * 24 * time.Hour)) {
		slog.Debug("Most recent match is already 3+ days in the future, skipping API call",
			"most_recent_date", mostRecentTimestamp)

		return true
	}

	return false
}
