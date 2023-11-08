package model

import (
	"crypto/ecdsa"
	"crypto/sha256"
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

func (Subkey) TableName() string {
	return "subkey"
}

// For `Secp256K1` : signature should be made by `personal_sign()`
// For `Secp256R1` : signature should be made by ECDSA w/ SHA256 under P-256 curve
func (subkey *Subkey) ValidateSignature(payload string, signature []byte) error {
	switch subkey.Algorithm {
	case types.SubkeyAlgorithms.Secp256K1:
		{
			pk, err := crypto.StringToSecp256k1Pubkey(subkey.PublicKey)
			if err != nil {
				return xerrors.Errorf("when deserializing subkey: %w", err)
			}
			return crypto.ValidatePersonalSignature(payload, signature, pk)
		}
	case types.SubkeyAlgorithms.Secp256R1:
		{
			pk, err := crypto.StringToSecp256r1Pubkey(subkey.PublicKey)
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
		return xerrors.Errorf("algorithm not supported: %s", subkey.Algorithm)
	}
}
