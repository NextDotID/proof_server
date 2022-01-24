package model

import (
	"crypto/ecdsa"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"

	"golang.org/x/xerrors"
	"gorm.io/datatypes"

	"github.com/nextdotid/proof-server/types"
	"github.com/nextdotid/proof-server/util/crypto"
	"github.com/nextdotid/proof-server/validator"
)

//  ProofChain is a chain of a persona's proof modification log.
type ProofChain struct {
	ID               int64          `gorm:"primarykey"`
	CreatedAt        time.Time      `gorm:"column:created_at"`
	Action           types.Action   `gorm:"index;not null"`
	Persona          string         `gorm:"index;not null"`
	Identity         string         `gorm:"index;not null"`
	Platform         types.Platform `gorm:"index;not null"`
	Location         string         `gorm:"not null"`
	Signature        string         `gorm:"not null"`
	SignaturePayload string         `gorm:"column:signature_payload"`
	Extra            datatypes.JSON `gorm:"default:'{}'"`
	PreviousID       sql.NullInt64  `gorm:"index"`
	Previous         *ProofChain
}

func (ProofChain) TableName() string {
	return "proof_chains"
}

func (pc *ProofChain) Pubkey() *ecdsa.PublicKey {
	pubkey, err := crypto.StringToPubkey(pc.Persona)
	if err != nil {
		return nil
	}
	return pubkey
}

// Apply applies current ProofChain modification to Proof model.
func (pc *ProofChain) Apply() (err error) {
	switch pc.Action {
	case types.Actions.Create:
		return pc.createProof()
	case types.Actions.Delete:
		return pc.deleteProof()
	default:
		return xerrors.Errorf("unknown action: %s", string(pc.Action))
	}
}

func (pc *ProofChain) createProof() (err error) {
	proof_found := Proof{
		Persona:      pc.Persona,
		Platform:     pc.Platform,
		Identity:     pc.Identity,
		Location:     pc.Location,
	}
	proof_create := &Proof{
		ProofChainID: pc.ID,
		Persona:      pc.Persona,
		Platform:     pc.Platform,
		Identity:     pc.Identity,
		Location:     pc.Location,
	}
	tx := DB.FirstOrCreate(proof_create, proof_found)
	if tx.Error != nil {
		return xerrors.Errorf("%w", tx.Error)
	}

	return nil
}

func (pc *ProofChain) deleteProof() (err error) {
	tx := DB.Delete(&Proof{}, Proof{
		Persona:  pc.Persona,
		Platform: pc.Platform,
		Identity: pc.Identity,
		Location: pc.Location,
	})
	if tx.Error != nil {
		return xerrors.Errorf("%w", tx.Error)
	}
	return nil
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

// MarshalSignature converts []byte signature into string.
func MarshalSignature(signature []byte) string {
	return base64.StdEncoding.EncodeToString(signature)
}

func ProofChainFindLatest(persona string) (pc *ProofChain, err error) {
	pc = new(ProofChain)
	tx := DB.Where("persona = ?", MarshalPersona(persona)).Order("id DESC").Take(pc)
	if tx.Error != nil {
		if strings.Contains(tx.Error.Error(), "record not found") {
			return nil, nil
		}
		return nil, xerrors.Errorf("%w", tx.Error)
	}

	return pc, nil
}

func ProofChainFindBySignature(signature string) (pc *ProofChain, err error) {
	previous := &ProofChain{}
	tx := DB.Where("signature = ?", signature).Take(previous)
	if tx.Error != nil || previous.ID == int64(0) {
		return nil, xerrors.Errorf("error finding previous proof chain: %w", tx.Error)
	}

	return previous, nil
}

func ProofChainCreateFromValidator(validator *validator.Base) (pc *ProofChain, err error) {
	pc = &ProofChain{
		Action:           validator.Action,
		Persona:          MarshalPersona(validator.Pubkey),
		Identity:         strings.ToLower(validator.Identity), // TODO: exception may occur
		Platform:         validator.Platform,
		Location:         validator.ProofLocation,
		Signature:        MarshalSignature(validator.Signature),
		SignaturePayload: validator.SignaturePayload,
		Previous:         nil,
	}

	if validator.Previous != "" {
		previous, err := ProofChainFindBySignature(validator.Previous)
		if err != nil {
			return nil, xerrors.Errorf("%w", err)
		}

		pc.Previous = previous
	}

	if len(validator.Extra) != 0 {
		extra_json, err := json.Marshal(validator.Extra)
		if err != nil {
			return nil, xerrors.Errorf("%w", err)
		}
		pc.Extra = datatypes.JSON(extra_json)
	}

	tx := DB.Create(pc)
	if tx.Error != nil {
		return nil, xerrors.Errorf("%w", err)
	}

	return pc, nil
}
