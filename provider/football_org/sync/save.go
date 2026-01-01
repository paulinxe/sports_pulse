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

func SaveMatches(footballOrgMatches []api.FootballOrgMatch, competition entity.Competition, teamMapping map[uint]entity.Team) error {
	tx, err := db.DB.Begin()
	if err != nil {
		return fmt.Errorf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	for _, footballOrgMatch := range footballOrgMatches {
		match, err := convertToEntityMatch(footballOrgMatch, competition, teamMapping)
		if err != nil {
			slog.Error(err.Error())
			continue
		}

		// As a match may be rescheduled, we need to delete the existing match in case it already exists.
		if err := repository.DeleteByCanonicalID(match.CanonicalID, entity.FootballOrg); err != nil {
			slog.Error("Failed to delete match", "error", err, "match", match)
		}

		if err := repository.Save(tx, *match); err != nil {
			return fmt.Errorf("Failed to insert match: %v", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("Failed to commit transaction: %v", err)
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

	match := entity.NewMatch(
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

	return &match, nil
}
