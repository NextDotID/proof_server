package github

import (
	"testing"

	"github.com/google/uuid"
	"github.com/nextdotid/proof-server/types"
	"github.com/nextdotid/proof-server/util"
	"github.com/nextdotid/proof-server/util/crypto"
	"github.com/nextdotid/proof-server/validator"
	"github.com/stretchr/testify/assert"
)

const (
	test_pubkey = "0x03947957e8a8785b6520b96c1c0d70ae9cf59835eec18f9ac920bbf5733413366a"
)

func generate() Github {
	pubkey, _ := crypto.StringToPubkey(test_pubkey)
	created_at, _ := util.TimestampStringToTime("1647329002")
	return Github{
		Base: &validator.Base{
			Platform:      types.Platforms.Github,
			Previous:      "",
			Action:        types.Actions.Create,
			Pubkey:        pubkey,
			Identity:      "nykma",
			ProofLocation: "5b3acc09d25242950e4b7ea0ee707ada",
			CreatedAt:     created_at,
			Uuid:          uuid.MustParse("909ee81f-4c5e-4319-affa-90d95eca614d"),
		},
	}
}

func Test_Validate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		github := generate()
		err := github.Validate()
		assert.Nil(t, err)
	})

	t.Run("error if owner mismatch", func(t *testing.T) {
		github := generate()
		github.Identity = "foobar"

		err := github.Validate()
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "gist owner mismatch")
	})

	t.Run("error if gist is private", func(t *testing.T) {
		github := generate()
		github.ProofLocation = "a8acd06e99ae6baa4939300fc170446c"

		err := github.Validate()
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "not found or empty")
	})
}
