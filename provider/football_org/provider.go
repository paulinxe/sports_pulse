package football_org

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"provider/entity"
	"provider/football_org/sync"
	"time"
)

type Provider struct{}

func (p *Provider) ValidateCompetition(competition entity.Competition) error {
	if _, ok := CompetitionToFootballOrgID[competition]; !ok {
		return fmt.Errorf("Competition not handled by football_org provider: %d", competition)
	}
	return nil
}

func (p *Provider) FetchMatches(ctx context.Context, competition entity.Competition, from, to time.Time) ([]entity.Match, error) {
	competitionID := CompetitionToFootballOrgID[competition]
	matchesResponse, err := sync.FetchAPI(ctx, competitionID, from, to)
	if err != nil {
		return nil, err
	}

	// Convert all matches to entity.Match
	entityMatches := make([]entity.Match, 0, len(matchesResponse.Matches))
	for _, footballOrgMatch := range matchesResponse.Matches {
		match, err := sync.ConvertToEntityMatch(footballOrgMatch, competition, FootballOrgTeamMapping)
		if err != nil {
			// Log error but continue processing other matches
			slog.Error("Failed to convert match to entity", "error", err, "match_id", footballOrgMatch.ID)
			continue
		}
		entityMatches = append(entityMatches, *match)
	}

	return entityMatches, nil
}

func (p *Provider) SaveMatches(ctx context.Context, tx *sql.Tx, matches []entity.Match, competition entity.Competition) error {
	return sync.SaveMatches(ctx, tx, matches, p.GetProviderEntity())
}

func (p *Provider) HasInProgressMatches(matches []entity.Match) bool {
	for _, match := range matches {
		if match.Status == entity.InProgress {
			return true
		}
	}
	return false
}

func (p *Provider) GetProviderEntity() entity.Provider {
	return entity.FootballOrg
}
