package model

import (
	"encoding/base64"
	"testing"

	"github.com/nextdotid/proof-server/types"
	"github.com/nextdotid/proof-server/util/crypto"
	"github.com/nextdotid/proof-server/validator"
	"github.com/stretchr/testify/assert"
)

func Test_ProofFindLatest(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)
		pk, _ := crypto.StringToPubkey("0x028c3cda474361179d653c41a62f6bbb07265d535121e19aedf660da2924d0b1e3")

		proof := Proof{
			Persona:       "0x" + crypto.CompressedPubkeyHex(pk),
			Platform:      "twitter",
			Identity:      "yeiwb",
			Location:      "1469221200140574721",
			Signature:     "gMUJ75eewkdaNrFp7bafzckv9+rlW7rVaxkB7/sYzYgFdFltYG+gn0lYzVNgrAdHWZPmu2giwJniGG7HG9iNigE=",
		}
		tx := DB.Create(&proof)
		assert.Nil(t, tx.Error)
		assert.NotEqual(t, uint(0), proof.ID)

		proof_found, err := ProofFindLatest(proof.Persona)
		assert.Nil(t, err)
		assert.NotNil(t, proof_found)
		assert.Equal(t, proof.ID, proof_found.ID)

		proof_new := Proof{
			PreviousProof: proof.ID,
			Persona:       proof.Persona,
			Platform:      "keybase",
			Identity:      "nykma",
			Location:      "",
			Signature:     "gMUJ75eewkdaNrFp7bafzckv9+rlW7rVaxkB7/sYzYgFdFltYG+gn0lYzVNgrAdHWZPmu2giwJniGG7HG9iNigE=",
		}
		tx = DB.Create(&proof_new)
		assert.Nil(t, tx.Error)
		assert.NotEqual(t, uint(0), proof_new.ID)

		proof_found_new, err := ProofFindLatest(proof.Persona)
		assert.Nil(t, err)
		assert.NotNil(t, proof_found_new)
		assert.Equal(t, proof_new.ID, proof_found_new.ID)
	})

	t.Run("should return empty resuot", func(t *testing.T) {
		before_each(t)

		proof, err := ProofFindLatest("foobar")
		assert.Nil(t, err)
		assert.Nil(t, proof)
	})
}

func Test_ProofCreateFromValidator(t *testing.T) {
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
		proof, err := ProofCreateFromValidator(&v)
		assert.Nil(t, err)
		assert.Equal(t, "yeiwb", proof.Identity)
		assert.Equal(t, base64.StdEncoding.EncodeToString([]byte{1, 2, 3, 4}), proof.Signature)
		assert.Equal(t, uint(0), proof.PreviousProof)
		assert.Equal(t, MarshalPersona(pk), proof.Persona)
		assert.Equal(t, "{}", proof.Extra.String())
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
		proof, err := ProofCreateFromValidator(&v)
		assert.Nil(t, err)
		assert.Equal(t, types.Platforms.Ethereum, proof.Platform)
		assert.Equal(t, `{"wallet_signature": "0xTEST"}`, proof.Extra.String())
	})
}
