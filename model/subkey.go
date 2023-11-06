package model

import (
	"time"

	"github.com/nextdotid/proof_server/types"
)

type Subkey struct {
	ID        int64     `gorm:"primarykey"`
	CreatedAt time.Time `gorm:"not null"`
	Name      string    `gorm:"not null"`
	// Relying Party Identifier.
	RP_ID string `gorm:"rp_id"`

	Avatar    string                `gorm:"not null"`
	// Algorithm of this subkey
	Algorithm types.SubkeyAlgorithm `gorm:"algorithm"`
	// Public key of this subkey.
	PublicKey string                `gorm:"public_key"`
}
