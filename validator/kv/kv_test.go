package kv

import (
	"encoding/json"
	"testing"

	"github.com/nextdotid/proof-server/model"
	"github.com/nextdotid/proof-server/types"
	"github.com/nextdotid/proof-server/util/crypto"
	"github.com/nextdotid/proof-server/validator"
	"github.com/stretchr/testify/assert"
)

const (
	test_kv_pubkey = "0x03ec5ff37fe2e0b0ff4e934796f42450184fb0dbbda33ff436a0cf0632e6a4c499"
)

func before_each(t *testing.T) {
	Init()
}

func test_generate_kv_base(t *testing.T) (validator.Base, validator.IValidator) {
	pubkey, _ := crypto.StringToPubkey(test_kv_pubkey)
	payload := model.KVPatch{
		Set: map[string]interface{}{
			"this": "is",
			"a": []string{"test"},
		},
		Del: []string{"non.exist"},
	}
	text, _ := json.Marshal(payload)

	base := validator.Base{
		Platform:         types.Platforms.KV,
		Previous:         "",
		Action:           types.Actions.KV,
		Pubkey:           pubkey,
		Text:             string(text),
		Extra:            map[string]string{},
	}
	validator_factory := validator.PlatformFactories[types.Platforms.KV]
	validator := validator_factory(&base)
	validator.GenerateSignPayload()

	return base, validator
}

func Test_GenerateSignPayload(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)
		_, validator := test_generate_kv_base(t)
		payload := validator.GenerateSignPayload()
		assert.Contains(t, payload, "kv")
		assert.Contains(t, payload, "\"set\"")
		assert.Contains(t, payload, "\"del\"")
		assert.Contains(t, payload, "\"prev\":null")
		assert.Contains(t, payload, "\"this\":\"is\"")
	})
}

func Test_Validate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)
		base, old_validator := test_generate_kv_base(t)
		pk, sk := crypto.GenerateKeypair()

		base.Pubkey = pk
		signature, err := crypto.SignPersonal(
			[]byte(old_validator.GenerateSignPayload()),
			sk,
		)
		assert.Nil(t, err)
		base.Signature = signature

		new_validator := validator.PlatformFactories[types.Platforms.KV]
		assert.Nil(t, new_validator(&base).Validate())
	})
}
