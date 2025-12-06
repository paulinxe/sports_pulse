package entity

import (
    "encoding/binary"
    "encoding/hex"
    "log/slog"
    "time"

    "github.com/ethereum/go-ethereum/crypto"
)

type Match struct {
    ID              string
    Start           time.Time
    End             time.Time
    Status          string
    Provider        Provider
    ProviderMatchID string
    HomeTeamID      int
    AwayTeamID      int
    HomeTeamScore   int
    AwayTeamScore   int
}

func NewMatch(
    start string,
    end string,
    provider Provider,
    providerMatchID string,
    homeTeamID int,
    awayTeamID int,
    homeTeamScore int,
    awayTeamScore int,
    competition Competition,
) Match {
    startTime, err := time.Parse(time.RFC3339, start)
    if err != nil {
        // TODO: if we can't parse the date, we should log an error. we will manually need to set the start time of the match.
        slog.Warn("Failed to parse match date, using current time",
            "provider_match_id", providerMatchID,
            "start", start,
            "error", err)
        startTime = time.Now() // TODO: we need to set the start time of the match.
    }

    // TODO: avoid magic numbers
    endTime := startTime.Add(2 * time.Hour)

    // TODO: we need to store the competition
    return Match{
        ID:              generateMatchID(uint64(competition), uint64(homeTeamID), uint64(awayTeamID), startTime),
        Start:           startTime,
        End:             endTime,
        Status:          "pending",
        Provider:        provider,
        ProviderMatchID: providerMatchID,
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
