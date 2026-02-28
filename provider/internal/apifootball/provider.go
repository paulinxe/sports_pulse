package apifootball

import (
	"fmt"
	"net/http"
	"os"
	"provider/internal/entity"
	"provider/internal/repository"
)

const defaultBaseURL = "https://apiv3.apifootball.com/"

type Provider struct {
	client                   *http.Client
	baseURL                  string
	apiKey                   string
	matchRepository          *repository.MatchRepository
	reconciliationRepository *repository.ReconciliationRepository
}

// NewProvider creates an APIfootball provider. APIFOOTBALL_API_KEY must be set.
// Optional APIFOOTBALL_API_ENDPOINT overrides the default base URL (e.g. for testing).
func NewProvider(matchRepo *repository.MatchRepository, reconciliationRepo *repository.ReconciliationRepository) *Provider {
	baseURL := os.Getenv("APIFOOTBALL_API_ENDPOINT")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	// Ensure trailing slash for building URLs
	// TODO: not sure if we need this
	if baseURL != "" && baseURL[len(baseURL)-1] != '/' {
		baseURL += "/"
	}

	return &Provider{
		client:                   &http.Client{},
		baseURL:                  baseURL,
		apiKey:                   os.Getenv("APIFOOTBALL_API_KEY"),
		matchRepository:          matchRepo,
		reconciliationRepository: reconciliationRepo,
	}
}

func (p *Provider) ValidateCompetition(competition entity.Competition) error {
	if _, ok := CompetitionToAPIFootballID[competition]; !ok {
		return fmt.Errorf("competition not handled by apifootball provider: %d", competition)
	}

	return nil
}

func (p *Provider) GetProviderEntity() entity.Provider {
	return entity.APIFootball
}
