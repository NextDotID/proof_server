package model

import (
	"encoding/base64"
	"testing"

	"github.com/nextdotid/proof-server/types"
	"github.com/nextdotid/proof-server/util/crypto"
	"github.com/nextdotid/proof-server/validator"
	"github.com/stretchr/testify/assert"
)

func Test_ProofChainFindBySignature(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)
		pk, _ := crypto.StringToPubkey("0x028c3cda474361179d653c41a62f6bbb07265d535121e19aedf660da2924d0b1e3")

		proof := ProofChain{
			Persona:       "0x" + crypto.CompressedPubkeyHex(pk),
			Platform:      "twitter",
			Identity:      "yeiwb",
			Location:      "1469221200140574721",
			Signature:     "gMUJ75eewkdaNrFp7bafzckv9+rlW7rVaxkB7/sYzYgFdFltYG+gn0lYzVNgrAdHWZPmu2giwJniGG7HG9iNigE=",
		}
		tx := DB.Create(&proof)
		assert.Nil(t, tx.Error)
		assert.Nil(t, proof.Previous)
	})

	t.Run("should return empty result", func(t *testing.T) {
		before_each(t)

		proof, err := ProofChainFindBySignature("0xfoobar")
		assert.Nil(t, proof)
		assert.Contains(t, err.Error(), "record not found")
	})
}

func Test_ProofChainCreateFromValidator(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)
		pk, _ := crypto.StringToPubkey("0x028c3cda474361179d653c41a62f6bbb07265d535121e19aedf660da2924d0b1e3")

		v := validator.Base{
			Platform:      types.Platforms.Twitter,
			Action:        types.Actions.Create,
			Pubkey:        pk,
			Identity:      "yeiwb",
			ProofLocation: "1469221200140574721",
			Signature:     []byte{1, 2, 3, 4},
			Text:          "",
		}
		pc, err := ProofChainCreateFromValidator(&v)
		assert.Nil(t, err)
		assert.Equal(t, "yeiwb", pc.Identity)
		assert.Equal(t, base64.StdEncoding.EncodeToString(v.Signature), pc.Signature)
		assert.Nil(t, pc.Previous)
		assert.Equal(t, MarshalPersona(pk), pc.Persona)
		assert.Equal(t, "{}", pc.Extra.String())
	})

	t.Run("save extra", func(t *testing.T) {
		before_each(t)
		pk, _ := crypto.StringToPubkey("0x028c3cda474361179d653c41a62f6bbb07265d535121e19aedf660da2924d0b1e3")

		v := validator.Base{
			Platform:      types.Platforms.Ethereum,
			Action:        types.Actions.Create,
			Pubkey:        pk,
			Identity:      "0xWALLET_ADDRESS",
			ProofLocation: "",
			Signature:     []byte{1, 2, 3, 4},
			Text:          "",
			Extra: map[string]string{
				"wallet_signature": "0xTEST",
			},
		}
		pc, err := ProofChainCreateFromValidator(&v)
		assert.Nil(t, err)
		assert.Equal(t, types.Platforms.Ethereum, pc.Platform)
		assert.Equal(t, `{"wallet_signature": "0xTEST"}`, pc.Extra.String())
	})

	t.Run("with previous connected", func(t *testing.T) {
		before_each(t)
		pk, _ := crypto.StringToPubkey("0x028c3cda474361179d653c41a62f6bbb07265d535121e19aedf660da2924d0b1e3")
		v := validator.Base{
			Platform:      types.Platforms.Twitter,
			Action:        types.Actions.Create,
			Pubkey:        pk,
			Identity:      "yeiwb",
			ProofLocation: "1469221200140574721",
			Signature:     []byte{1, 2, 3, 4},
			Text:          "",
		}
		prev, err := ProofChainCreateFromValidator(&v)
		assert.Nil(t, err)

		v2 := validator.Base{
			Previous:      MarshalSignature(v.Signature),
			Platform:      types.Platforms.Twitter,
			Action:        types.Actions.Delete,
			Pubkey:        pk,
			Identity:      "yeiwb",
			ProofLocation: "1469221200140574721",
			Signature:     []byte{5, 6, 7, 8},
			Text:          "",
		}
		current, err := ProofChainCreateFromValidator(&v2)
		assert.Nil(t, err)
		assert.Equal(t, prev.ID, current.Previous.ID)
		assert.Equal(t, prev.ID, current.PreviousID.Int64)
	})

	t.Run("cannot connect to previous", func(t *testing.T) {
		before_each(t)

		pk, _ := crypto.StringToPubkey("0x028c3cda474361179d653c41a62f6bbb07265d535121e19aedf660da2924d0b1e3")
		v := validator.Base{
			Previous:      MarshalSignature([]byte{1, 2, 3, 4}),
			Platform:      types.Platforms.Twitter,
			Action:        types.Actions.Delete,
			Pubkey:        pk,
			Identity:      "yeiwb",
			ProofLocation: "1469221200140574721",
			Signature:     []byte{5, 6, 7, 8},
			Text:          "",
		}
		pc, err := ProofChainCreateFromValidator(&v)
		assert.Nil(t, pc)
		assert.Contains(t, err.Error(), "record not found")
	})
}

func Test_Apply(t *testing.T) {
	t.Run("create and delete", func(t *testing.T) {
		before_each(t)
		pk, _ := crypto.StringToPubkey("0x028c3cda474361179d653c41a62f6bbb07265d535121e19aedf660da2924d0b1e3")

		pc := ProofChain{
			Action:     types.Actions.Create,
			Persona:    MarshalPersona(pk),
			Identity:   "yeiwb",
			Location: "1469221200140574721",
			Platform:   types.Platforms.Twitter,
			Signature:  MarshalSignature([]byte{1, 2, 3, 4}),
		}
		tx := DB.Create(&pc)
		assert.Nil(t, tx.Error)

		err := pc.Apply()
		assert.Nil(t, err)

		proof_found := Proof{ProofChainID: pc.ID,}
		DB.First(&proof_found)

		assert.NotZero(t, proof_found.ID)
		assert.Equal(t, pc.Location, proof_found.Location)
		assert.Equal(t, pc.Identity, proof_found.Identity)

		// Duplicated apply
		err = pc.Apply()
		assert.Nil(t, err)
		var count int64
		DB.Model(&Proof{}).Where("proof_chain_id = ?", pc.ID).Count(&count)
		assert.Equal(t, int64(1), count)

		// Delete
		pc_delete := ProofChain{
			Action:     types.Actions.Delete,
			Persona:    MarshalPersona(pk),
			Identity:   "yeiwb",
			Location: "1469221200140574721",
			Platform:   types.Platforms.Twitter,
			Signature:  MarshalSignature([]byte{1, 2, 3, 4}),
		}
		pc_delete.Apply()
		DB.Model(&Proof{}).Where("proof_chain_id = ?", pc.ID).Count(&count)
		assert.Equal(t, int64(0), count)
	})
}
