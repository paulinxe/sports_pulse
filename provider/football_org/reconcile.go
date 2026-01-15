package football_org

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"provider/entity"
	"provider/football_org/api"
	"provider/repository"
	"time"
)

const RECONCILE_INTERVAL = -24 * time.Hour

// The purpose of this reconcile is to update the matches that are already in the database.
// We only check matches that should have ended by now and 1 day ago and are still in Pending status.
// TODO: For those matches that should have ended but remained in Pending for more than 1 day, we need a new watchdog.
// The idea is to run it periodically to ensure the matches are up to date. (TODO: determine the frequency)
func Reconcile(timeoutPerMatch time.Duration) error {
	// TODO: do we need some mutex here? if we plan to run this function in parallel, we need to use a mutex to avoid race conditions.
	// Or even if its single threaded, at the moment we don't have a Context Timeout for the function itself.
	if timeoutPerMatch == 0 {
		timeoutPerMatch = 15 * time.Second
	}

	now := time.Now()
	matches, err := repository.FindMatchesToReconcile(context.Background(), entity.FootballOrg, now.Add(RECONCILE_INTERVAL), now)
	if err != nil {
		return err
	}

	slog.Debug("Found matches to reconcile", "number_of_matches", len(matches))

	for _, match := range matches {
		slog.Debug("Reconciling match", "match_id", match.ID)

		// Create context with timeout for HTTP request and database operation
		ctx, cancel := context.WithTimeout(context.Background(), timeoutPerMatch)
		defer cancel()

		apiResponse, err := api.GetOne(ctx, fmt.Sprintf("/matches/%s", match.ID))
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				slog.Error("Context timeout while fetching match from API", "match_id", match.ID, "timeout", timeoutPerMatch)
			} else {
				slog.Error("Failed to fetch match from API", "match_id", match.ID, "error", err)
			}

			continue
		}

		// We know for sure we will get 1 match only as the endpoint responds with a JSON object or 400 when the match is not found
		if apiResponse.Status != "FINISHED" {
			slog.Debug("Match is not finished, skipping reconciliation", "match_id", match.ID)
			// TODO: depending on the status, we may need to add a delay to the next reconciliation
			continue
		}

		if err := repository.FinishMatch(ctx, match.ID, apiResponse.Score.FullTime.Home, apiResponse.Score.FullTime.Away); err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				slog.Error("Context timeout while finishing match in database", "match_id", match.ID, "timeout", timeoutPerMatch)
			} else {
				slog.Error("Failed to finish match", "match_id", match.ID, "error", err)
			}

			continue
		}

		slog.Debug("Finished match", "match_id", match.ID, "home_team_score", apiResponse.Score.FullTime.Home, "away_team_score", apiResponse.Score.FullTime.Away)
	}

	return nil
}
