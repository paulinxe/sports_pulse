package service

import (
	"context"
	"fmt"
	"log/slog"
	"provider/internal/entity"
	"provider/internal/football_org"
	"provider/internal/repository"
	"strings"
	"time"
)

const SYNC_CONTEXT_TIMEOUT = 15 * time.Second
const STALE_MATCH_THRESHOLD = 6 * time.Hour

// Clock provides an interface for getting the current time, allowing for time mocking in tests
type Clock interface {
	Now() time.Time
}

// SystemClock implements Clock using the system time.
type SystemClock struct{}

func (s SystemClock) Now() time.Time {
	return time.Now().UTC()
}

// SyncProvider defines the interface for provider-specific sync operations
type SyncProvider interface {
	ValidateCompetition(competition entity.Competition) error
	FetchMatches(ctx context.Context, competition entity.Competition, from, to time.Time) ([]entity.Match, error)
	SaveMatches(ctx context.Context, matches []entity.Match)
	GetProviderEntity() entity.Provider
}

// ReconcileProvider extends SyncProvider with the ability to fetch a single match by ID (for reconciliation)
type ReconcileProvider interface {
	SyncProvider
	FetchMatchByID(ctx context.Context, providerMatchID string) (*entity.Match, error)
}

// Sync queries matches for a natural day period (00:00:00 to 23:59:59 UTC) and only inserts
// matches with FINISHED or AWARDED status. It implements catch-up logic to advance day-by-day
// until reaching today, and checks for in-progress matches to avoid advancing too early.
func Sync(repositories *repository.Repositories, provider string, competition string, clock Clock) error {
	competitionEntity := entity.Competition(0)
	switch strings.ToLower(competition) {
	case "la_liga":
		competitionEntity = entity.LaLiga
	case "premier_league":
		competitionEntity = entity.PremierLeague
	default:
		return fmt.Errorf("unknown competition: %s", competition)
	}

	// Get the provider implementation
	var syncProvider SyncProvider
	switch strings.ToLower(provider) {
	case "football_org":
		syncProvider = football_org.NewProvider(repositories.Match, repositories.Reconciliation)
	default:
		return fmt.Errorf("unknown provider: %s", provider)
	}

	if err := syncProvider.ValidateCompetition(competitionEntity); err != nil {
		return err
	}

	// Create context with 15-second timeout for HTTP requests and database operations
	ctx, cancel := context.WithTimeout(context.Background(), SYNC_CONTEXT_TIMEOUT)
	defer cancel()

	now := clock.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	// Determine which day to query
	queryDate, err := getQueryDate(ctx, repositories.SyncState, competitionEntity, &today, syncProvider.GetProviderEntity())
	if err != nil {
		return fmt.Errorf("failed to get query date: %w", err)
	}

	// Query for the natural day (00:00:00 to 23:59:59 UTC)
	from := queryDate                   // 00:00:00 UTC
	to := queryDate.Add(24 * time.Hour) // 00:00:00 UTC next day (exclusive)

	matchesResponse, err := syncProvider.FetchMatches(ctx, competitionEntity, from, to)
	if err != nil {
		// On error, log and retry the same day next time (don't advance)
		return fmt.Errorf("failed to fetch matches: %w", err)
	}

	filteredMatches, err := filterStaleMatches(ctx, repositories.Reconciliation, matchesResponse, syncProvider.GetProviderEntity(), clock.Now())
	if err != nil {
		return fmt.Errorf("failed to filter stale matches: %w", err)
	}

	syncProvider.SaveMatches(ctx, filteredMatches)

	var nextSyncDate time.Time
	if hasInProgressMatches(filteredMatches) {
		// Matches still in progress: stay on current day, will retry in 30 min
		nextSyncDate = queryDate
		slog.Debug("Matches still in progress, staying on current day", "date", queryDate)
	} else if queryDate.Before(today) {
		// All matches finished and we're catching up: advance by 1 day
		nextSyncDate = queryDate.Add(24 * time.Hour)
		slog.Debug("All matches finished, advancing sync date by 1 day", "from", queryDate, "to", nextSyncDate)
	} else {
		// Already on today and no matches in progress: stay on today (will query again in 30 min)
		nextSyncDate = today
		slog.Debug("Staying on today", "date", today)
	}

	if err := repositories.SyncState.UpdateLastSyncedDate(ctx, competitionEntity, syncProvider.GetProviderEntity(), nextSyncDate); err != nil {
		return fmt.Errorf("failed to update last synced date: %w", err)
	}

	return nil
}

func getQueryDate(ctx context.Context, syncStateRepo *repository.SyncStateRepository, competition entity.Competition, today *time.Time, provider entity.Provider) (time.Time, error) {
	lastSyncedDate, err := syncStateRepo.GetLastSyncedDate(ctx, competition, provider)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get last synced date: %w", err)
	}

	if lastSyncedDate == nil {
		return *today, nil
	}

	// Truncate lastSyncedDate to day boundary (UTC)
	lastSyncedDay := time.Date(lastSyncedDate.Year(), lastSyncedDate.Month(), lastSyncedDate.Day(), 0, 0, 0, 0, time.UTC)

	if lastSyncedDay.After(*today) {
		// Future date (shouldn't happen, but handle it)
		slog.Info("Sync date is in the future, updating sync date to today", "last_synced_day", lastSyncedDay, "today", today)
		return *today, nil
	}

	return lastSyncedDay, nil
}

// filterStaleMatches identifies matches that started 6+ hours ago and are not finished,
// moves them to the reconciliation queue, and returns a filtered list without stale matches.
// This catches both in-progress matches that have been running too long, and matches that
// should have started but haven't reached a finished state (e.g., still Pending).
func filterStaleMatches(
	ctx context.Context,
	reconciliation *repository.ReconciliationRepository,
	matches []entity.Match,
	provider entity.Provider,
	now time.Time,
) ([]entity.Match, error) {
	var filteredMatches []entity.Match
	staleThreshold := now.Add(-STALE_MATCH_THRESHOLD)

	for _, match := range matches {
		// Check if match is not finished and started more than 6 hours ago
		if match.Status != entity.Finished && match.Start.Before(staleThreshold) {
			if err := reconciliation.SaveToReconciliationQueue(ctx, match.ProviderMatchID, provider); err != nil {
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

// hasInProgressMatches checks if there is at least one match with Status != Finished
func hasInProgressMatches(matches []entity.Match) bool {
	for _, match := range matches {
		if match.Status != entity.Finished {
			return true
		}
	}

	return false
}
