package football_org

import (
	"context"
	"fmt"
	"log/slog"
	"provider/db"
	"provider/entity"
	"provider/football_org/sync"
	"provider/repository"
	"time"
)

const SYNC_CONTEXT_TIMEOUT = 15 * time.Second
const SYNC_INTERVAL = 3 * 24 * time.Hour

// This Sync function is intended to run once a day to fetch the latest matches and insert them into the database.
// It will skip the sync if the most recent match is already 3+ days in the future.
// In reconcile.go we take care of updating the matches that are already in the database.
func Sync(competition entity.Competition) error {
	if err := validateCompetition(competition); err != nil {
		return err
	}

	// Create context with 15-second timeout for HTTP requests and database operations
	ctx, cancel := context.WithTimeout(context.Background(), SYNC_CONTEXT_TIMEOUT)
	defer cancel()

	lastSyncedDate, err := repository.GetLastSyncedDate(ctx, competition, entity.FootballOrg)
	if err != nil {
		return fmt.Errorf("Failed to get last synced date: %w", err)
	}

	var from time.Time
	if lastSyncedDate == nil {
		// Use today as default if no sync state exists (first sync)
		from = time.Now()
	} else {
		from = *lastSyncedDate
	}

	if shouldSkipSync(from) {
		return nil
	}

	to := from.Add(SYNC_INTERVAL)
	competitionID := CompetitionToFootballOrgID[competition]
	matchesResponse, err := sync.FetchAPI(ctx, competitionID, from, to)
	if err != nil {
		return err
	}

	// Open transaction for both saving matches and updating sync state
	tx, err := db.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("Failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if err := sync.SaveMatches(ctx, tx, matchesResponse.Matches, competition, FootballOrgTeamMapping); err != nil {
		return err
	}

	var nextSyncAt time.Time
	if len(matchesResponse.Matches) == 0 {
		// No matches found: advance by 1 day from the start of the checked range
		nextSyncAt = from.Add(24 * time.Hour)
		slog.Debug("No matches found, advancing sync date by 1 day", "new_sync_date", nextSyncAt)
	} else {
		slog.Debug(fmt.Sprintf("Successfully inserted %d matches into database", len(matchesResponse.Matches)))

		// Matches found: update to end of checked range to avoid re-checking
		nextSyncAt = to
		slog.Debug("Matches found, updating sync date to end of range", "new_sync_date", nextSyncAt)
	}

	if err := repository.UpdateLastSyncedDate(ctx, tx, competition, entity.FootballOrg, nextSyncAt); err != nil {
		return fmt.Errorf("Failed to update last synced date: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("Failed to commit transaction: %w", err)
	}

	return nil
}

func validateCompetition(competition entity.Competition) error {
	if _, ok := CompetitionToFootballOrgID[competition]; !ok {
		return fmt.Errorf("unknown competition: %d", competition)
	}

	return nil
}

func shouldSkipSync(syncTimestamp time.Time) bool {
	now := time.Now()
	if syncTimestamp.After(now.Add(3 * 24 * time.Hour)) {
		slog.Debug("Sync timestamp is already 3+ days in the future, skipping API call",
			"sync_timestamp", syncTimestamp)

		return true
	}

	return false
}
