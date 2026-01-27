package entity

import (
	"github.com/google/uuid"
)

// Match is a signed match ready to be broadcast to the chain.
// CanonicalID is the bytes32 match id (stored as hex in DB).
// Start is YYYYMMDD as uint32 for the contract.
// Signature is the signer's EIP-712 signature (hex in DB, bytes for contract).
type Match struct {
	ID            uuid.UUID
	CanonicalID   string
	CompetitionID int32
	HomeTeamID    int32
	AwayTeamID    int32
	HomeTeamScore int32
	AwayTeamScore int32
	Start         uint32
	Signature     []byte
}
