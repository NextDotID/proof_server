package model

import (
	"testing"

	"github.com/nextdotid/proof-server/types"
	"github.com/nextdotid/proof-server/util/crypto"
	"github.com/nextdotid/proof-server/validator/twitter"
	"github.com/stretchr/testify/assert"
	// "gorm.io/datatypes"
)

func Test_Proof_Validate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)
		twitter.Init()

		pk, _ := crypto.StringToPubkey("0x028c3cda474361179d653c41a62f6bbb07265d535121e19aedf660da2924d0b1e3")
		pc := ProofChain{
			Action:    types.Actions.Create,
			Persona:   MarshalPersona(pk),
			Identity:  "yeiwb",
			Location:  "1469221200140574721",
			Platform:  types.Platforms.Twitter,
			Signature: "gMUJ75eewkdaNrFp7bafzckv9+rlW7rVaxkB7/sYzYgFdFltYG+gn0lYzVNgrAdHWZPmu2giwJniGG7HG9iNigE=",
		}
		tx := DB.Create(&pc)
		assert.Nil(t, tx.Error)

		err := pc.Apply()
		assert.Nil(t, err)

		proof := new(Proof)
		DB.Where("location = ?", pc.Location).Preload("ProofChain").Find(proof)
		assert.NotEqual(t, proof.ID, 0)

		result, err := proof.Revalidate()
		assert.True(t, result)
		assert.Nil(t, err)
	})

	t.Run("failure", func(t *testing.T) {
		before_each(t)
		twitter.Init()

		pk, _ := crypto.StringToPubkey("0x028c3cda474361179d653c41a62f6bbb07265d535121e19aedf660da2924d0b1e3")
		pc := ProofChain{
			Action:    types.Actions.Create,
			Persona:   MarshalPersona(pk),
			Identity:  "yeiwb",
			Location:  "1469221200140574720",
			Platform:  types.Platforms.Twitter,
			Signature: "gMUJ75eewkdaNrFp7bafzckv9+rlW7rVaxkB7/sYzYgFdFltYG+gn0lYzVNgrAdHWZPmu2giwJniGG7HG9iNigE=",
		}
		tx := DB.Create(&pc)
		assert.Nil(t, tx.Error)

		err := pc.Apply()
		assert.Nil(t, err)

		proof := new(Proof)
		DB.Where("location = ?", pc.Location).Preload("ProofChain").Find(proof)
		assert.NotEqual(t, proof.ID, 0)

		result, err := proof.Revalidate()
		assert.False(t, result)
		t.Logf("%s", err.Error())
	})
}
