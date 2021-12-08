package model

import (
	"time"

	"golang.org/x/xerrors"
	"gorm.io/gorm"
)

type Platform string

// PlatformAvailable is a list of all current supported platforms, DO NOT MODIFY IT IN RUNTIME.
var PlatformAvailable = struct {
	Twitter  Platform
	Keybase  Platform
	Ethereum Platform
}{
	Twitter:  "twitter",
	Keybase:  "keybase",
	Ethereum: "ethereum",
}

type Proof struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time

	PreviousProof uint     `gorm:"index;not null"`
	// Persona is public key of user persona (string, /0x[0-9a-f]{130}/)
	Persona       string   `gorm:"index;not null"`
	Platform      Platform `gorm:"index;not null"`
	Identity      string   `gorm:"index;not null"`
	Link          string   `gorm:"not null"`
	Signature     string   `gorm:"not null"`
}

func (Proof) TableName() string {
	return "proof"
}

// Previous returns previous proof of self.
func (proof *Proof) Previous() (prevProof *Proof, err error) {
	if proof.PreviousProof == uint(0) {
		return nil, nil
	}

	previous := new(Proof)
	tx := DB.First(previous, proof.PreviousProof)

	if tx.Error != nil {
		return nil, xerrors.Errorf("%w", tx.Error)
	}
	return previous, nil
}

// ProofFindLatest finds latest Proof in the chain by given persona pubkey.
func ProofFindLatest(persona string) (proof *Proof, err error) {
	proof = new(Proof)
	// FIXME: make this correct by link iteration.
	tx := DB.Where("persona = ?", persona).Last(proof)
	if tx.Error != nil {
		return nil, xerrors.Errorf("%w", tx.Error)
	}
	return proof, nil
}
