package model

import (
	"crypto/ecdsa"
	"encoding/base64"
	"strings"
	"time"

	"golang.org/x/xerrors"

	"github.com/nextdotid/proof-server/types"
	"github.com/nextdotid/proof-server/util/crypto"
	"github.com/nextdotid/proof-server/validator"
)

type Proof struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time

	PreviousProof uint `gorm:"index;not null"`
	// Persona is public key of user persona (string, /0x[0-9a-f]{130}/)
	Persona   string         `gorm:"index;not null"`
	Platform  types.Platform `gorm:"index;not null"`
	Identity  string         `gorm:"index;not null"`
	Location  string         `gorm:"not null"`
	Signature string         `gorm:"not null"`
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

func (proof *Proof) Pubkey() *ecdsa.PublicKey {
	pubkey, err := crypto.StringToPubkey(proof.Persona)
	if err != nil {
		return nil
	}
	return pubkey
}

// MarshalPersona accepts *ecdsa.Pubkey | string type of pubkey,
// returns a string to be stored into DB.
func MarshalPersona(persona interface{}) string {
	switch p := persona.(type) {
	case *ecdsa.PublicKey:
		return "0x" + crypto.CompressedPubkeyHex(p)
	case string:
		pubkey, err := crypto.StringToPubkey(p)
		if err != nil {
			return ""
		}
		return MarshalPersona(pubkey)
	default:
		return ""
	}
}

// ProofFindLatest finds latest Proof in the chain by given persona pubkey.
func ProofFindLatest(persona string) (proof *Proof, err error) {
	proof = new(Proof)

	// FIXME: make this procedure correct by linklist iteration.
	tx := DB.Where("persona = ?", MarshalPersona(persona)).Last(proof)
	if tx.Error != nil {
		if strings.Contains(tx.Error.Error(), "record not found") {
			return nil, nil
		}
		return nil, xerrors.Errorf("%w", tx.Error)
	}
	return proof, nil
}

func ProofCreateFromValidator(validator *validator.Base) (proof *Proof, err error) {
	proof = &Proof{
		PreviousProof: 0,
		Persona:       MarshalPersona(validator.Pubkey),
		Platform:      validator.Platform,
		Identity:      validator.Identity,
		Location:      validator.ProofLocation,
		Signature:     base64.StdEncoding.EncodeToString(validator.Signature),
	}
	if validator.Previous != "" {
		previous := &Proof{Signature: validator.Previous}
		tx := DB.First(previous)
		if tx.Error != nil || previous.ID == uint(0) {
			return nil, xerrors.Errorf("error finding previous proof: %w", tx.Error)
		}

		proof.PreviousProof = previous.ID
	}

	tx := DB.Create(proof)
	if tx.Error != nil {
		return nil, xerrors.Errorf("%w", tx.Error)
	}

	return proof, nil
}
