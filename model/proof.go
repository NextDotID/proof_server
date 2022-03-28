package model

import (
	"time"

	"github.com/nextdotid/proof-server/types"
	"github.com/nextdotid/proof-server/validator"
	"golang.org/x/xerrors"
)

// Proof is final proof state of a user (persona).
type Proof struct {
	ID            int64 `gorm:"primarykey"`
	CreatedAt     time.Time
	LastCheckedAt time.Time
	IsValid       bool
	InvalidReason string

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

func FindAllProofByPersona(persona any) (proofs []Proof, err error) {
	marshaled_persona := MarshalPersona(persona)
	proofs = make([]Proof, 0)
	tx := DB.Model(&Proof{}).Where("persona = ?", marshaled_persona).Find(&proofs)
	if tx.Error != nil {
		return nil, xerrors.Errorf("error when finding proofs: %w", err)
	}
	return proofs, nil
}

// Revalidate validates current proof, will update `IsValid` and
// `LastCheckedAt`. Must be used after `DB.Preload("ProofChain")`.
func (proof *Proof) Revalidate() (result bool, err error) {
	v, err := proof.ProofChain.RestoreValidator()
	if err != nil {
		return false, xerrors.Errorf("error when restoring validator: %w", err)
	}

	iv := validator.BaseToInterface(v)
	if iv == nil {
		return false, xerrors.Errorf("unknown platform: %s", string(proof.Platform))
	}

	err = iv.Validate()
	if err != nil {
		proof.touchValid(false, err.Error())
		return false, xerrors.Errorf("validate failed: %w", err)
	}

	proof.touchValid(true, "")
	return true, nil
}

func (proof *Proof) touchValid(result bool, reason string) {
	proof.LastCheckedAt = time.Now()
	proof.IsValid = result
	proof.InvalidReason = reason
	DB.Save(proof)
}
