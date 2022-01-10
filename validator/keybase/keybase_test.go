package keybase

import (
	"testing"

	"github.com/nextdotid/proof-server/config"
	"github.com/nextdotid/proof-server/types"
	mycrypto "github.com/nextdotid/proof-server/util/crypto"
	"github.com/nextdotid/proof-server/validator"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func before_each(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	config.Init("../../config/config.test.json")
}

func generate() Keybase {
	pubkey, _ := mycrypto.StringToPubkey("0x033d2c5c16bc24ced47619bd3471cef57c8ea8ecce9268700286d61de0d9f3f2dd")
	return Keybase{
		Base: &validator.Base{
			Previous: "",
			Action:   "create",
			Pubkey:   pubkey,
			Identity: "nykma",
			Platform: types.Platforms.Keybase,
		},
	}
}

func Test_GeneratePostPayload(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)

		kb := generate()
		result := kb.GeneratePostPayload()
		assert.Contains(t, result, "To validate")
		assert.Contains(t, result, mycrypto.CompressedPubkeyHex(kb.Pubkey))
		assert.Contains(t, result, "%%SIG_BASE64%%")
		t.Logf("%+v", result)
	})
}

func Test_Validate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)

		kb := generate()
		kb.Identity = "NYKma"
		assert.Nil(t, kb.Validate())
		assert.Greater(t, len(kb.Signature), 10)
		assert.Equal(t, "nykma", kb.Identity)
	})
}
