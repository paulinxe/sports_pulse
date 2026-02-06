package entity

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
)

type Match struct {
	ID              uuid.UUID
	CanonicalID     string
	Start           time.Time
	End             time.Time
	Status          MatchStatus
	Provider        Provider
	ProviderMatchID string
	CompetitionID   Competition
	HomeTeamID      Team
	AwayTeamID      Team
	HomeTeamScore   uint
	AwayTeamScore   uint
}

type MatchStatus int

const (
	_ MatchStatus = iota
	Pending
	InProgress
	Finished
)

func NewMatch(
	start time.Time,
	provider Provider,
	providerMatchID string,
	homeTeamID Team,
	awayTeamID Team,
	homeTeamScore uint,
	awayTeamScore uint,
	competition Competition,
	status MatchStatus,
) (Match, error) {
	endTime := start.Add(2 * time.Hour)

	if status == 0 {
		status = Pending
	}

	canonicalID, err := generateMatchID(competition, homeTeamID, awayTeamID, start)
	if err != nil {
		return Match{}, fmt.Errorf("failed to generate match ID: %w", err)
	}

	return Match{
		ID:              uuid.New(),
		CanonicalID:     canonicalID,
		Start:           start,
		End:             endTime,
		Status:          status,
		Provider:        provider,
		ProviderMatchID: providerMatchID,
		CompetitionID:   competition,
		HomeTeamID:      homeTeamID,
		AwayTeamID:      awayTeamID,
		HomeTeamScore:   homeTeamScore,
		AwayTeamScore:   awayTeamScore,
	}, nil
}

func generateMatchID(compId Competition, homeTeamId Team, awayTeamId Team, matchDay time.Time) (string, error) {
	// Equivalent to abi.encodePacked(uint32, uint32, uint32, uint32)
	var packed []byte

	packed = append(packed, uint32ToBytes(uint32(compId))...)
	packed = append(packed, uint32ToBytes(uint32(homeTeamId))...)
	packed = append(packed, uint32ToBytes(uint32(awayTeamId))...)

	// Convert date to uint32 in format YYYYMMDD
	dateStr := matchDay.Format("20060102")
	dateUint64, err := strconv.ParseUint(dateStr, 10, 32)
	if err != nil {
		return "", fmt.Errorf("failed to parse date string %s: %w", dateStr, err)
	}
	packed = append(packed, uint32ToBytes(uint32(dateUint64))...)

	hash := crypto.Keccak256(packed)

	return hex.EncodeToString(hash), nil
}

func uint32ToBytes(v uint32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, v)
	return b
}
