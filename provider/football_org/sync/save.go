package sync

import (
	"fmt"
	"log/slog"
	"provider/db"
	"provider/entity"
	"provider/football_org/api"
	"provider/repository"
	"time"
)

func SaveMatches(footballOrgMatches []api.FootballOrgMatch, competition entity.Competition, teamMapping map[int]entity.Team) error {
	tx, err := db.DB.Begin()
	if err != nil {
		slog.Error("Failed to begin transaction", "error", err)
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	for _, footballOrgMatch := range footballOrgMatches {
		match, err := convertToEntityMatch(footballOrgMatch, competition, teamMapping)
		if err != nil {
			slog.Error("Failed to convert match", "error", err, "match_id", footballOrgMatch.ID)
			continue
		}

		if match == nil {
			continue // Already logged in convertToEntityMatch
		}

		if err := repository.Save(tx, *match); err != nil {
			// TODO: if a match was moved, here we will have a duplicate key sql error.
			// in this case, we need to remove the existing match and insert the new one.
			slog.Error("Failed to insert match", "error", err, "match", match)
			continue
		}
	}

	if err := tx.Commit(); err != nil {
		slog.Error("Failed to commit transaction", "error", err)
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

func convertToEntityMatch(footballOrgMatch api.FootballOrgMatch, competition entity.Competition, teamMapping map[int]entity.Team) (*entity.Match, error) {
	homeTeamID, ok := teamMapping[footballOrgMatch.HomeTeam.ID]
	if !ok {
		slog.Error("Failed to map home team ID, skipping match",
			"external_team_id", footballOrgMatch.HomeTeam.ID,
			"match_id", footballOrgMatch.ID)
		return nil, nil
	}

	awayTeamID, ok := teamMapping[footballOrgMatch.AwayTeam.ID]
	if !ok {
		slog.Error("Failed to map away team ID, skipping match",
			"external_team_id", footballOrgMatch.AwayTeam.ID,
			"match_id", footballOrgMatch.ID)
		return nil, nil
	}

	startTime, err := time.Parse(time.RFC3339, footballOrgMatch.UTCDate)
	if err != nil {
		slog.Error("Failed to parse match date", "error", err, "match_id", footballOrgMatch.ID)
		return nil, err
	}

	match := entity.NewMatch(
		startTime,
		entity.FootballOrg,
		fmt.Sprintf("%d", footballOrgMatch.ID),
		homeTeamID,
		awayTeamID,
		footballOrgMatch.Score.FullTime.Home,
		footballOrgMatch.Score.FullTime.Away,
		competition,
	)

	return &match, nil
}
