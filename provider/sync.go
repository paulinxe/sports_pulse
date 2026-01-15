package main

import (
	"fmt"
	"provider/entity"
	"provider/football_org"
	"strings"
)

func Sync(provider string, competition string) error {
	competitionEntity := entity.Competition(0)
	switch strings.ToLower(competition) {
	case "la_liga":
		competitionEntity = entity.LaLiga
	default:
		return fmt.Errorf("Unknown competition: %s", competition)
	}

	switch strings.ToLower(provider) {
	case "football_org":
		if err := football_org.Sync(competitionEntity); err != nil {
			return fmt.Errorf("Failed to sync Football Data API: %w", err)
		}
	default:
		return fmt.Errorf("Unknown provider: %s", provider)
	}

	return nil
}
