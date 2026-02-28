package football_org

import (
	"fmt"
	"os"
	"provider/internal/entity"

	"github.com/paulinxe/go-football-data"
)

// Provider implements sync.SyncProvider for the football_org data source.
type Provider struct {
	client *football_data.Client
}

func NewProvider() *Provider {
	opts := []football_data.Option[football_data.Client]{}
	if customEndpoint := os.Getenv("FOOTBALL_ORG_API_ENDPOINT"); customEndpoint != "" {
		// Mainly used for testing but could be useful if we need to use a different endpoint for some reason.
		opts = append(opts, football_data.WithBaseURL(customEndpoint))
	}

	return &Provider{
		client: football_data.New(os.Getenv("FOOTBALL_ORG_API_KEY"), opts...),
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
