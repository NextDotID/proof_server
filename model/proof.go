package model

import (
	"fmt"
	"time"

	"github.com/nextdotid/proof_server/types"
	"github.com/nextdotid/proof_server/validator"
	"golang.org/x/xerrors"
)

// EXPIRED_IN is the time after which a proof is considered expired and should perform revalidate.
const EXPIRED_IN = time.Hour * 24 * 3

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
	AltID    string         `gorm:"column:alt_id;index"`
	Location string         `gorm:"not null"`
}

func (Proof) TableName() string {
	return "proof"
}

func FindAllProofByPersona(persona any, orderBy string) (proofs []Proof, err error) {
	marshaled_persona := MarshalPersona(persona)
	proofs = make([]Proof, 0)
	tx := ReadOnlyDB.Model(&Proof{}).Where("persona = ?", marshaled_persona).Order(orderBy).Find(&proofs)
	if tx.Error != nil {
		return nil, xerrors.Errorf("error when finding proofs: %w", err)
	}
	return proofs, nil
}

// IsOutdated returns true if proof is outdated and should do a revalidate.
func (proof *Proof) IsOutdated() bool {
	return proof.LastCheckedAt.Add(EXPIRED_IN).Before(time.Now())
}

// Revalidate validates current proof, will update `IsValid` and
// `LastCheckedAt`. Must be used after `DB.Preload("ProofChain")`.
func (proof *Proof) Revalidate() (err error) {
	v, err := proof.ProofChain.RestoreValidator()
	if err != nil || v == nil {
		return xerrors.Errorf("restoring validator: %w", err)
	}

	iv := validator.BaseToInterface(v)
	if iv == nil {
		return xerrors.Errorf("unknown platform: %s", string(proof.Platform))
	}

	err = iv.Validate()
	if err != nil {
		proof.touchValid(err.Error(), iv.GetAltID())
		return xerrors.Errorf("validate failed: %w", err)
	}

	proof.touchValid("", iv.GetAltID())
	// TODO: need to update `identity` and `alt_id` here.
	return nil
}

func (proof *Proof) touchValid(reason, altID string) {
	fmt.Printf("AltID: %s\n", altID)
	proof.LastCheckedAt = time.Now()
	proof.IsValid = (reason == "")
	proof.InvalidReason = reason
	if altID != "" {
		proof.AltID = altID
	}

	DB.Save(proof)
}
