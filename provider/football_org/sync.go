package football_org

import (
	"context"
	"fmt"
	"log/slog"
	"provider/db"
	"provider/entity"
	"provider/football_org/api"
	"provider/football_org/sync"
	"provider/repository"
	"time"
)

const SYNC_CONTEXT_TIMEOUT = 15 * time.Second

// Sync queries matches for a natural day period (00:00:00 to 23:59:59 UTC) and only inserts
// matches with FINISHED or AWARDED status. It implements catch-up logic to advance day-by-day
// until reaching today, and checks for in-progress matches to avoid advancing too early.
func Sync(competition entity.Competition) error {
	if err := validateCompetition(competition); err != nil {
		return err
	}

	// Create context with 15-second timeout for HTTP requests and database operations
	ctx, cancel := context.WithTimeout(context.Background(), SYNC_CONTEXT_TIMEOUT)
	defer cancel()

	// Always work in UTC
	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	// Determine which day to query
	queryDate, err := getQueryDate(&ctx, competition, &today)
	if err != nil {
		return fmt.Errorf("Failed to get query date: %w", err)
	}

	// Query for the natural day (00:00:00 to 23:59:59 UTC)
	from := queryDate                   // 00:00:00 UTC
	to := queryDate.Add(24 * time.Hour) // 00:00:00 UTC next day (exclusive)

	competitionID := CompetitionToFootballOrgID[competition]
	matchesResponse, err := sync.FetchAPI(ctx, competitionID, from, to)
	if err != nil {
		// On error, log and retry the same day next time (don't advance)
		slog.Error("Failed to fetch matches from API", "error", err, "date", queryDate)
		return fmt.Errorf("Failed to fetch matches: %w", err)
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

	// Check if any matches are still in progress
	hasInProgress := hasInProgressMatches(matchesResponse.Matches)

	var nextSyncDate time.Time
	if hasInProgress {
		// Matches still in progress: stay on current day, will retry in 30 min
		// TODO: we should have a way to detect we are waiting too much time for a match to finish.
		// Then we would need to add this match to a "dead-letter" queue for reconciliation and advance one day.
		nextSyncDate = queryDate
		slog.Debug("Matches still in progress, staying on current day", "date", queryDate)
	} else if queryDate.Before(today) {
		// All matches finished and we're catching up: advance by 1 day
		// TODO: if are close to 00:00 (next day), we should check if we have matches that didn't start yet.
		// If this is the case, in order to not lose that match, we should not advance the sync date and query again in 30 min.
		// This could happen if there is a delay in the API response.
		// Another option could be to move that match to a "dead-letter" queue for reconciliation and advance one day.
		nextSyncDate = queryDate.Add(24 * time.Hour)
		slog.Debug("All matches finished, advancing sync date by 1 day", "from", queryDate, "to", nextSyncDate)
	} else {
		// Already on today and no matches in progress: stay on today (will query again in 30 min)
		nextSyncDate = today
		slog.Debug("Staying on today", "date", today)
	}

	if err := repository.UpdateLastSyncedDate(ctx, tx, competition, entity.FootballOrg, nextSyncDate); err != nil {
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

func getQueryDate(ctx *context.Context, competition entity.Competition, today *time.Time) (time.Time, error) {
	lastSyncedDate, err := repository.GetLastSyncedDate(*ctx, competition, entity.FootballOrg)
	if err != nil {
		return time.Time{}, fmt.Errorf("Failed to get last synced date: %w", err)
	}

	if lastSyncedDate == nil {
		return *today, nil
	}

	// Truncate lastSyncedDate to day boundary (UTC)
	lastSyncedDay := time.Date(lastSyncedDate.Year(), lastSyncedDate.Month(), lastSyncedDate.Day(), 0, 0, 0, 0, time.UTC)

	if lastSyncedDay.After(*today) {
		// Future date (shouldn't happen, but handle it)
		// TODO: instead of erroring, we could update the sync date to today and continue.
		return time.Time{}, fmt.Errorf("sync date is in the future: %s", lastSyncedDay)
	}

	return lastSyncedDay, nil
}

// hasInProgressMatches checks if any matches from the API response are still in progress.
func hasInProgressMatches(matches []api.FootballOrgMatch) bool {
	for _, match := range matches {
		if match.IsInProgress() {
			return true
		}
	}
	return false
}