package main

import (
	"context"
	"fmt"
	"log/slog"
	"provider/entity"
	"provider/football_org"
	"provider/repository"
	"strings"
	"time"
)

const SYNC_CONTEXT_TIMEOUT = 15 * time.Second
const STALE_MATCH_THRESHOLD = 6 * time.Hour

// SyncProvider defines the interface for provider-specific sync operations
type SyncProvider interface {
	ValidateCompetition(competition entity.Competition) error
	FetchMatches(ctx context.Context, competition entity.Competition, from, to time.Time) ([]entity.Match, error)
	SaveMatches(ctx context.Context, matches []entity.Match)
	HasInProgressMatches(matches []entity.Match) bool
	GetProviderEntity() entity.Provider
}

// Sync queries matches for a natural day period (00:00:00 to 23:59:59 UTC) and only inserts
// matches with FINISHED or AWARDED status. It implements catch-up logic to advance day-by-day
// until reaching today, and checks for in-progress matches to avoid advancing too early.
func Sync(provider string, competition string) error {
	competitionEntity := entity.Competition(0)
	switch strings.ToLower(competition) {
	case "la_liga":
		competitionEntity = entity.LaLiga
	case "premier_league":
		competitionEntity = entity.PremierLeague
	default:
		return fmt.Errorf("Unknown competition: %s", competition)
	}

	// Get the provider implementation
	var syncProvider SyncProvider
	switch strings.ToLower(provider) {
	case "football_org":
		syncProvider = &football_org.Provider{}
	default:
		return fmt.Errorf("Unknown provider: %s", provider)
	}

	if err := syncProvider.ValidateCompetition(competitionEntity); err != nil {
		return err
	}

	// Create context with 15-second timeout for HTTP requests and database operations
	ctx, cancel := context.WithTimeout(context.Background(), SYNC_CONTEXT_TIMEOUT)
	defer cancel()

	// Always work in UTC
	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	// Determine which day to query
	queryDate, err := getQueryDate(&ctx, competitionEntity, &today, syncProvider.GetProviderEntity())
	if err != nil {
		return fmt.Errorf("Failed to get query date: %w", err)
	}

	// Query for the natural day (00:00:00 to 23:59:59 UTC)
	from := queryDate                   // 00:00:00 UTC
	to := queryDate.Add(24 * time.Hour) // 00:00:00 UTC next day (exclusive)

	matchesResponse, err := syncProvider.FetchMatches(ctx, competitionEntity, from, to)
	if err != nil {
		// On error, log and retry the same day next time (don't advance)
		slog.Error("Failed to fetch matches from API", "error", err, "date", queryDate)
		return fmt.Errorf("Failed to fetch matches: %w", err)
	}

	syncProvider.SaveMatches(ctx, matchesResponse)

	// Filter stale in-progress matches and move them to reconciliation queue
	// Get fresh now value for accurate stale match detection
	currentTime := time.Now().UTC()
	filteredMatches, err := filterStaleMatches(ctx, matchesResponse, syncProvider.GetProviderEntity(), currentTime)
	if err != nil {
		return fmt.Errorf("Failed to filter stale matches: %w", err)
	}

	// Check if any matches are still in progress (after filtering stale ones)
	hasInProgress := syncProvider.HasInProgressMatches(filteredMatches)

	var nextSyncDate time.Time
	if hasInProgress {
		// Matches still in progress: stay on current day, will retry in 30 min
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

	if err := repository.UpdateLastSyncedDate(ctx, competitionEntity, syncProvider.GetProviderEntity(), nextSyncDate); err != nil {
		return fmt.Errorf("Failed to update last synced date: %w", err)
	}

	return nil
}

func getQueryDate(ctx *context.Context, competition entity.Competition, today *time.Time, provider entity.Provider) (time.Time, error) {
	lastSyncedDate, err := repository.GetLastSyncedDate(*ctx, competition, provider)
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

// filterStaleMatches identifies matches that started 6+ hours ago and are still in-progress,
// moves them to the reconciliation queue, and returns a filtered list without stale matches.
func filterStaleMatches(
	ctx context.Context,
	matches []entity.Match,
	provider entity.Provider,
	now time.Time,
) ([]entity.Match, error) {
	var filteredMatches []entity.Match
	staleThreshold := now.Add(-STALE_MATCH_THRESHOLD)

	for _, match := range matches {
		// Check if match is in-progress and started more than 6 hours ago
		if match.Status == entity.InProgress && match.Start.Before(staleThreshold) {
			if err := repository.SaveToReconciliationQueue(ctx, match.ProviderMatchID, provider); err != nil {
				// Log error but continue - don't fail the entire sync
				slog.Error("Failed to add stale match to reconciliation queue",
					"provider_match_id", match.ProviderMatchID,
					"provider", provider,
					"error", err)

				continue
			}

			slog.Info("Moved stale match to reconciliation queue",
				"provider_match_id", match.ProviderMatchID,
				"provider", provider,
			)

			continue
		}

		filteredMatches = append(filteredMatches, match)
	}

	return filteredMatches, nil
}
