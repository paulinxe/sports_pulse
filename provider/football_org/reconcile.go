package football_org

import (
	"context"
	"fmt"
	"provider/entity"
	"provider/repository"
	"time"
	"provider/football_org/api"
	"log/slog"
)

const RECONCILE_INTERVAL = -24 * time.Hour
const RECONCILE_CONTEXT_TIMEOUT = 15 * time.Second

// The purpose of this reconcile is to update the matches that are already in the database.
// We only check matches that should have ended by now and 1 day ago and are still in Pending status.
// TODO: For those matches that should have ended but remained in Pending for more than 1 day, we need a new watchdog.
// The idea is to run it periodically to ensure the matches are up to date. (TODO: determine the frequency)
func Reconcile() error {
	// TODO: do we need some mutex here?
	now := time.Now()
	matches, err := repository.FindMatchesToReconcile(context.Background(), entity.FootballOrg, now.Add(RECONCILE_INTERVAL), now)
	if err != nil {
		return err
	}

	slog.Debug("Found matches to reconcile", "number_of_matches", len(matches))

	for _, match := range matches {
		slog.Debug("Reconciling match", "match_id", match.ProviderMatchID)

		// Create context with 15-second timeout for HTTP request and database operation
		ctx, cancel := context.WithTimeout(context.Background(), RECONCILE_CONTEXT_TIMEOUT)

		apiResponse, err := api.GetOne(ctx, fmt.Sprintf("/matches/%s", match.ProviderMatchID))
		if err != nil {
			cancel()
			slog.Error(err.Error())
			continue
		}

		// We know for sure we will get 1 match only as the endpoint responds with a JSON object or 400 when the match is not found
		if apiResponse.Status != "FINISHED" {
			cancel()
			slog.Debug("Match is not finished, skipping reconciliation", "match_id", match.ProviderMatchID)
			// TODO: depending on the status, we may need to add a delay to the next reconciliation
			continue
		}

		if err := repository.FinishMatch(ctx, match.ID, apiResponse.Score.FullTime.Home, apiResponse.Score.FullTime.Away); err != nil {
			cancel()
			slog.Error("Failed to finish match", "match_id", match.ProviderMatchID, "error", err)
			continue
		}

		cancel()
		slog.Debug("Finished match", "match_id", match.ProviderMatchID, "home_team_score", apiResponse.Score.FullTime.Home, "away_team_score", apiResponse.Score.FullTime.Away)
	}

	return nil
}