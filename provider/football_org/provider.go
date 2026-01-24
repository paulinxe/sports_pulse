package football_org

import (
	"fmt"
	"provider/entity"
)

type Provider struct{}

func (p *Provider) ValidateCompetition(competition entity.Competition) error {
	if _, ok := CompetitionToFootballOrgID[competition]; !ok {
		return fmt.Errorf("Competition not handled by football_org provider: %d", competition)
	}
	return nil
}

func (p *Provider) GetProviderEntity() entity.Provider {
	return entity.FootballOrg
}
