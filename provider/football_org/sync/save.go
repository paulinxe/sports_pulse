package sync

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"provider/entity"
	"provider/football_org/api"
	"provider/repository"
	"time"
)

func SaveMatches(ctx context.Context, tx *sql.Tx, footballOrgMatches []api.FootballOrgMatch, competition entity.Competition, teamMapping map[uint]entity.Team) error {
	for _, footballOrgMatch := range footballOrgMatches {
		match, err := convertToEntityMatch(footballOrgMatch, competition, teamMapping)
		if err != nil {
			slog.Error(err.Error())
			continue
		}

		// As a match may be rescheduled, we need to delete the existing match in case it already exists.
		if err := repository.DeleteByCanonicalID(ctx, tx, match.CanonicalID, entity.FootballOrg); err != nil {
			slog.Error("Failed to delete match", "error", err, "match", match)
		}

		if err := repository.Save(ctx, tx, *match); err != nil {
			// TODO: a single match fail should not fail the entire sync.
			// We should have a "dead-letter" queue for failed matches so we reconcile each of them individually later.
			return fmt.Errorf("Failed to insert match: %w", err)
		}
	}

	return nil
}

func convertToEntityMatch(footballOrgMatch api.FootballOrgMatch, competition entity.Competition, teamMapping map[uint]entity.Team) (*entity.Match, error) {
	homeTeamID, ok := teamMapping[footballOrgMatch.HomeTeam.ID]
	if !ok {
		return nil, fmt.Errorf("Failed to map home team ID (%d), skipping match (%d)",
			footballOrgMatch.HomeTeam.ID,
			footballOrgMatch.ID,
		)
	}

	awayTeamID, ok := teamMapping[footballOrgMatch.AwayTeam.ID]
	if !ok {
		return nil, fmt.Errorf("Failed to map away team ID (%d), skipping match (%d)",
			footballOrgMatch.AwayTeam.ID,
			footballOrgMatch.ID,
		)
	}

	startTime, err := time.Parse(time.RFC3339, footballOrgMatch.UTCDate)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse match date (%s), skipping match (%d)",
			footballOrgMatch.UTCDate,
			footballOrgMatch.ID,
		)
	}

	status := entity.Pending
	if footballOrgMatch.Status == "FINISHED" {
		status = entity.Finished
	}

	match, err := entity.NewMatch(
		startTime,
		entity.FootballOrg,
		fmt.Sprintf("%d", footballOrgMatch.ID),
		homeTeamID,
		awayTeamID,
		footballOrgMatch.Score.FullTime.Home,
		footballOrgMatch.Score.FullTime.Away,
		competition,
		status,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create match: %w", err)
	}

	return &match, nil
}
