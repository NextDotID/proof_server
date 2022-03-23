package model

import (
	"encoding/base64"
	"strings"
	"testing"
	"time"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	"github.com/nextdotid/proof-server/types"
	"github.com/nextdotid/proof-server/util"
	"github.com/nextdotid/proof-server/util/crypto"
	"github.com/nextdotid/proof-server/validator"
	"github.com/nextdotid/proof-server/validator/ethereum"
	"github.com/nextdotid/proof-server/validator/twitter"
	"github.com/stretchr/testify/assert"
)

func Test_Proof_Revalidate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)
		twitter.Init()

		pk, _ := crypto.StringToPubkey("0x04666b700aeb6a6429f13cbb263e1bc566cd975a118b61bc796204109c1b351d19b7df23cc47f004e10fef41df82bad646b027578f8881f5f1d2f70c80dfcd8031")
		orig_created_at, _ := util.TimestampStringToTime("1647503071")
		pc := ProofChain{
			Action:    types.Actions.Create,
			Persona:   MarshalPersona(pk),
			Identity:  "yeiwb",
			Location:  "1504363098328924163",
			Platform:  types.Platforms.Twitter,
			Signature: "D8i0UOXKrHJ23zCQe6USZDrw7fOjwm4R/eVX0AZXKgomynWWm+Px4Y7I1wtbsHwKj0t9psFqm87EnM93DXOmhwE=",
			Uuid:      "c6fa1483-1bad-4f07-b661-678b191ab4b3",
			CreatedAt: orig_created_at,
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
			Uuid:      uuid.New().String(),
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
