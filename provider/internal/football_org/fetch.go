package football_org

import (
	"context"
	"fmt"
	"log/slog"
	"provider/internal/entity"
	"time"

	"github.com/paulinxe/go-football-data"
	"github.com/paulinxe/go-football-data/types"
)

func (p *Provider) FetchMatchByID(ctx context.Context, providerMatchID string) (*entity.Match, error) {
	footballOrgMatch := types.Match{}
	err := p.client.GetMatch(ctx, providerMatchID, &footballOrgMatch)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch match %s: %w", providerMatchID, err)
	}

	competition, ok := FootballOrgIDToCompetition[footballOrgMatch.Competition.ID]
	if !ok {
		return nil, fmt.Errorf("unknown competition ID %d for match %s", footballOrgMatch.Competition.ID, providerMatchID)
	}

	match, err := convertToEntityMatch(footballOrgMatch, competition, FootballOrgTeamMapping)
	if err != nil {
		return nil, fmt.Errorf("failed to convert match %s: %w", providerMatchID, err)
	}

	return match, nil
}

func (p *Provider) FetchMatches(ctx context.Context, competition entity.Competition, from, to time.Time) ([]entity.Match, error) {
	competitionID := CompetitionToFootballOrgID[competition]
	matchesResponse := types.MatchesList{}
	err := p.client.GetMatches(ctx, competitionID, football_data.MatchesFilter{
		DateFrom: &from,
		DateTo:   &to,
	}, &matchesResponse)
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
func convertToEntityMatch(footballOrgMatch types.Match, competition entity.Competition, teamMapping map[uint]entity.Team) (*entity.Match, error) {
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
		return nil, fmt.Errorf("failed to parse match date (%s): %w, skipping match (%d)",
			footballOrgMatch.UTCDate,
			err,
			footballOrgMatch.ID,
		)
	}

	// Map API status to entity status
	status := entity.Pending
	if isInFinalStatus(&footballOrgMatch) {
		status = entity.Finished
	} else if isInProgress(&footballOrgMatch) {
		status = entity.InProgress
	}

	homeTeamScore := uint(0)
	if footballOrgMatch.Score.FullTime.Home != nil {
		homeTeamScore = *footballOrgMatch.Score.FullTime.Home
	}
	awayTeamScore := uint(0)
	if footballOrgMatch.Score.FullTime.Away != nil {
		awayTeamScore = *footballOrgMatch.Score.FullTime.Away
	}

	match, err := entity.NewMatch(
		startTime,
		entity.FootballOrg,
		fmt.Sprintf("%d", footballOrgMatch.ID),
		homeTeamID,
		awayTeamID,
		homeTeamScore,
		awayTeamScore,
		competition,
		status,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create match: %w", err)
	}

	return match, nil
}

// The happy path is to have a match in status FINISHED.
// If a match gets cancelled and never gets played, it will be in status AWARDED.
func isInFinalStatus(match *types.Match) bool {
	return match.Status == "FINISHED" || match.Status == "AWARDED"
}

// Statuses: IN_PLAY, PAUSED, SUSPENDED indicate matches that are actively in progress.
// Note: TIMED, SCHEDULED are not in-progress (match hasn't started yet).
func isInProgress(match *types.Match) bool {
	inProgressStatuses := map[string]bool{
		"IN_PLAY":   true,
		"PAUSED":    true,
		"SUSPENDED": true,
	}
	return inProgressStatuses[match.Status]
}
