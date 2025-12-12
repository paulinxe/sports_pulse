package entity

import "github.com/google/uuid"

type ReconciliableMatch struct {
	ID uuid.UUID
	ProviderMatchID string
	HomeTeamScore uint
	AwayTeamScore uint
}