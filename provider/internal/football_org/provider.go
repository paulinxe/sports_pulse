package football_org

import (
	"fmt"
	"provider/internal/entity"
	"provider/internal/repository"
)

// Provider implements sync.SyncProvider for the football_org data source.
type Provider struct {
	matchRepository          *repository.MatchRepository
	reconciliationRepository *repository.ReconciliationRepository
}

func NewProvider(matchRepo *repository.MatchRepository, reconciliationRepo *repository.ReconciliationRepository) *Provider {
	return &Provider{
		matchRepository:          matchRepo,
		reconciliationRepository: reconciliationRepo,
	}
}

func (p *Provider) ValidateCompetition(competition entity.Competition) error {
	if _, ok := CompetitionToFootballOrgID[competition]; !ok {
		return fmt.Errorf("competition not handled by football_org provider: %d", competition)
	}
	return nil
}

func (p *Provider) GetProviderEntity() entity.Provider {
	return entity.FootballOrg
}
