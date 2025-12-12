package entity

import (
    "encoding/binary"
    "encoding/hex"
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
	Processing
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
) Match {
    endTime := start.Add(2 * time.Hour)

    return Match{
        ID:              uuid.New(),
        CanonicalID:     generateMatchID(uint64(competition), uint64(homeTeamID), uint64(awayTeamID), start),
        Start:           start,
        End:             endTime,
        Status:          Pending,
        Provider:        provider,
        ProviderMatchID: providerMatchID,
        CompetitionID:   competition,
        HomeTeamID:      homeTeamID,
        AwayTeamID:      awayTeamID,
        HomeTeamScore:   homeTeamScore,
        AwayTeamScore:   awayTeamScore,
    }
}

func generateMatchID(compId, homeTeamId, awayTeamId uint64, matchDay time.Time) string {
    // TODO: maybe 64 bytes is too much for the id. we should use a shorter id.
    // Equivalent to abi.encodePacked(uint64, uint64, uint64, string)
    var packed []byte

    packed = append(packed, uint64ToBytes(compId)...)
    packed = append(packed, uint64ToBytes(homeTeamId)...)
    packed = append(packed, uint64ToBytes(awayTeamId)...)
    packed = append(packed, []byte(matchDay.Format("20060102"))...)

    hash := crypto.Keccak256(packed)

    return hex.EncodeToString(hash)
}

func uint64ToBytes(v uint64) []byte {
    b := make([]byte, 8)
    binary.BigEndian.PutUint64(b, v)
    return b
}
