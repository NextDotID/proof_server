package github

import (
	"testing"

	"github.com/nextdotid/proof-server/types"
	"github.com/nextdotid/proof-server/util/crypto"
	"github.com/nextdotid/proof-server/validator"
	"github.com/stretchr/testify/assert"
)

const (
	test_pubkey = "0x03ec5ff37fe2e0b0ff4e934796f42450184fb0dbbda33ff436a0cf0632e6a4c499"
)

func generate() Github {
	pubkey, _ := crypto.StringToPubkey(test_pubkey)
	return Github{
		Base: &validator.Base{
			Platform:      types.Platforms.Github,
			Previous:      "",
			Action:        types.Actions.Create,
			Pubkey:        pubkey,
			Identity:      "nykma",
			ProofLocation: "5b3acc09d25242950e4b7ea0ee707ada",
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
