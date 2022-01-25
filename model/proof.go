package model

import (
	"time"

	"github.com/nextdotid/proof-server/types"
)

// Proof is final proof state of a user (persona).
type Proof struct {
	ID            int64 `gorm:"primarykey"`
	CreatedAt     time.Time
	LastCheckedAt time.Time
	IsValid         bool

	ProofChainID int64 `gorm:"index"`
	ProofChain   ProofChain
	// Persona is public key of user persona (string, /0x[0-9a-f]{130}/)
	Persona  string         `gorm:"index;not null"`
	Platform types.Platform `gorm:"index;not null"`
	Identity string         `gorm:"index;not null"`
	Location string         `gorm:"not null"`
}

func (Proof) TableName() string {
	return "proof"
}
