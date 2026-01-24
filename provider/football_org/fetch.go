package football_org

import (
	"context"
	"fmt"
	"log/slog"
	"provider/entity"
	"provider/football_org/api"
	"time"
)

func (p *Provider) FetchMatches(ctx context.Context, competition entity.Competition, from, to time.Time) ([]entity.Match, error) {
	competitionID := CompetitionToFootballOrgID[competition]
	matchesResponse, err := api.GetList(ctx, competitionID, from, to)
	if err != nil {
		return nil, err
	}

	// Convert all matches to entity.Match
	entityMatches := make([]entity.Match, 0, len(matchesResponse.Matches))
	for _, footballOrgMatch := range matchesResponse.Matches {
		match, err := convertToEntityMatch(footballOrgMatch, competition, FootballOrgTeamMapping)
		if err != nil {
			// Log error but continue processing other matches
			slog.Error("Failed to convert match to entity", "error", err, "match_id", footballOrgMatch.ID)
			continue
		}
		entityMatches = append(entityMatches, *match)
	}

	return entityMatches, nil
}

// convertToEntityMatch converts a FootballOrgMatch to entity.Match, handling all statuses.
func convertToEntityMatch(footballOrgMatch api.FootballOrgMatch, competition entity.Competition, teamMapping map[uint]entity.Team) (*entity.Match, error) {
	homeTeamID, ok := teamMapping[footballOrgMatch.HomeTeam.ID]
	if !ok {
		return nil, fmt.Errorf("failed to map home team ID (%d), skipping match (%d)",
			footballOrgMatch.HomeTeam.ID,
			footballOrgMatch.ID,
		)
	}

	awayTeamID, ok := teamMapping[footballOrgMatch.AwayTeam.ID]
	if !ok {
		return nil, fmt.Errorf("failed to map away team ID (%d), skipping match (%d)",
			footballOrgMatch.AwayTeam.ID,
			footballOrgMatch.ID,
		)
	}

	startTime, err := time.Parse(time.RFC3339, footballOrgMatch.UTCDate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse match date (%s), skipping match (%d)",
			footballOrgMatch.UTCDate,
			footballOrgMatch.ID,
		)
	}

	// Map API status to entity status
	status := entity.Pending
	if footballOrgMatch.IsInFinalStatus() {
		status = entity.Finished
	} else if footballOrgMatch.IsInProgress() {
		status = entity.InProgress
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
