package model

import (
	"testing"

	"github.com/google/uuid"
	"github.com/nextdotid/proof_server/types"
	"github.com/nextdotid/proof_server/util"
	"github.com/nextdotid/proof_server/util/crypto"
	"github.com/nextdotid/proof_server/validator/twitter"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
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
		require.NoError(t, tx.Error)

		err := pc.Apply()
		require.Nil(t, err)

		proof := new(Proof)
		DB.Where("location = ?", pc.Location).Preload("ProofChain").Find(proof)
		require.NotEqual(t, proof.ID, 0)

		require.NoError(t, proof.Revalidate())
		require.NotEmpty(t, proof.AltID, "should update AltID when revalidating")
	})

	// t.Run("failure", func(t *testing.T) {
	// 	before_each(t)
	// 	twitter.Init()

	// 	pk, _ := crypto.StringToPubkey("0x028c3cda474361179d653c41a62f6bbb07265d535121e19aedf660da2924d0b1e3")
	// 	pc := ProofChain{
	// 		Action:    types.Actions.Create,
	// 		Persona:   MarshalPersona(pk),
	// 		Identity:  "yeiwb",
	// 		Location:  "1469221200140574720",
	// 		Platform:  types.Platforms.Twitter,
	// 		Signature: "gMUJ75eewkdaNrFp7bafzckv9+rlW7rVaxkB7/sYzYgFdFltYG+gn0lYzVNgrAdHWZPmu2giwJniGG7HG9iNigE=",
	// 		Uuid:      uuid.New().String(),
	// 	}
	// 	tx := DB.Create(&pc)
	// 	require.Nil(t, tx.Error)

	// 	err := pc.Apply()
	// 	require.Nil(t, err)

	// 	proof := new(Proof)
	// 	DB.Where("location = ?", pc.Location).Preload("ProofChain").Find(proof)
	// 	require.NotEqual(t, proof.ID, 0)

	// 	require.Error(t, proof.Revalidate())
	// })
}

func Test_FindAllProofByPersona(t *testing.T) {
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
		require.Nil(t, tx.Error)
		err := pc.Apply()
		require.Nil(t, err)

		pc2 := ProofChain{
			Action:    types.Actions.Create,
			Persona:   MarshalPersona(pk),
			Identity:  "0x....",
			Location:  "",
			Platform:  types.Platforms.Ethereum,
			Signature: "D8i0UOXKrHJ23zCQe6USZDrw7fOjwm4R/eVX0AZXKgomynWWm+Px4Y7I1wtbsHwKj0t9psFqm87EnM93DXOmhwE=",
			Uuid:      uuid.New().String(),
			CreatedAt: orig_created_at,
		}
		tx = DB.Create(&pc2)
		require.Nil(t, tx.Error)
		err = pc2.Apply()
		require.Nil(t, err)

		// 2 records should be created
		var count int64
		DB.Model(&Proof{}).Count(&count)
		require.Equal(t, int64(2), count)

		proofs, err := FindAllProofByPersona(pk, "id desc")
		require.Nil(t, err)
		require.Equal(t, 2, len(proofs))
		result_types := lo.Map(proofs, func(proof Proof, _index int) types.Platform {
			return proof.Platform
		})
		require.Contains(t, result_types, types.Platforms.Twitter)
		require.Contains(t, result_types, types.Platforms.Ethereum)
	})
}
