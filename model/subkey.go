package model

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/json"
	"math/big"
	"time"

	"github.com/nextdotid/proof_server/types"
	"github.com/nextdotid/proof_server/util/crypto"
	"golang.org/x/xerrors"
)

type Subkey struct {
	ID        int64     `gorm:"primarykey"`
	CreatedAt time.Time `gorm:"not null"`
	Name      string    `gorm:"not null"`
	// Relying Party Identifier.
	RP_ID string `gorm:"rp_id"`

	Avatar string `gorm:"not null"`
	// Algorithm of this subkey
	Algorithm types.SubkeyAlgorithm `gorm:"algorithm"`
	// Public key of this subkey.
	PublicKey string `gorm:"public_key"`
}

type subkeySignPayload struct {
	Avatar    string `json:"avatar"`
	Algorithm string `json:"algorithm"`
	PublicKey string `json:"public_key"`
	RP_ID     string `json:"rp_id"`
}

func (Subkey) TableName() string {
	return "subkey"
}

// `self` doesn't needed to be stored in DB.
func (self *Subkey) GenerateSignPayload() (payload string, err error) {
	if self.RP_ID == "" {
		return "", xerrors.Errorf("rp_id is empty")
	}
	avatarPK, err := crypto.StringToSecp256k1Pubkey(self.Avatar)
	if err != nil {
		return "", xerrors.Errorf("when parsing avatar public key: %w", err)
	}
	switch self.Algorithm {
	case types.SubkeyAlgorithms.Secp256R1: {
		_, err = crypto.StringToSecp256r1Pubkey(self.PublicKey)
	}
	case types.SubkeyAlgorithms.Secp256K1: {
		_, err = crypto.StringToSecp256k1Pubkey(self.PublicKey)
	}
	}
	if err != nil {
		return "", xerrors.Errorf("when parsing subkey public key: %w", err)
	}
	payloadStruct := subkeySignPayload {
		Avatar: crypto.CompressedPubkeyHex(avatarPK),
		Algorithm: string(self.Algorithm),
		PublicKey: self.PublicKey,
		RP_ID: self.RP_ID,
	}

	payloadBytes, err := json.Marshal(payloadStruct)
	if err != nil {
		return "", xerrors.Errorf("when marshal JSON: %w", err)
	}
	return string(payloadBytes), nil
}

// For `Secp256K1` : signature should be made by `personal_sign()`.
// For `Secp256R1` : signature should be made by ECDSA w/ SHA256 under P-256 curve
func (self *Subkey) ValidateSignature(payload string, signature []byte) error {
	switch self.Algorithm {
	case types.SubkeyAlgorithms.Secp256K1:
		{
			pk, err := crypto.StringToSecp256k1Pubkey(self.PublicKey)
			if err != nil {
				return xerrors.Errorf("when deserializing subkey: %w", err)
			}
			return crypto.ValidatePersonalSignature(payload, signature, pk)
		}
	case types.SubkeyAlgorithms.Secp256R1:
		{
			pk, err := crypto.StringToSecp256r1Pubkey(self.PublicKey)
			if err != nil {
				return xerrors.Errorf("when deserializing subkey: %w", err)
			}
			hash := sha256.Sum256([]byte(payload))
			r := new(big.Int).SetBytes(signature[:32])
			s := new(big.Int).SetBytes(signature[32:])
			if ecdsa.Verify(pk, hash[:], r, s) {
				return nil
			} else {
				return xerrors.New("signature validation failed")
			}

		}
	default:
		return xerrors.Errorf("algorithm not supported: %s", self.Algorithm)
	}
}

func (self *Subkey) Save(signature []byte) (id int64, err error) {
	signPayload, err := self.GenerateSignPayload()
	if err != nil {
		return 0, xerrors.Errorf("when generationg sign payload: %w", err)
	}
	if err := self.ValidateSignature(signPayload, signature); err != nil {
		return 0, xerrors.Errorf("when validating signature: %w", err)
	}

	self.CreatedAt = time.Now()
	tx := DB.Create(self)
	if tx.Error != nil {
		return 0, xerrors.Errorf("when saving record: %w", err)
	}

	return self.ID, nil
}
