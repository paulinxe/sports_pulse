package entity

import (
    "github.com/google/uuid"
)

type MatchStatus int
const (
    _ MatchStatus = iota
    Pending
	Processing
	Finished
	Signed
)

type Match struct {
    ID              uuid.UUID
    CanonicalID     string
    HomeTeamScore   uint
    AwayTeamScore   uint
}