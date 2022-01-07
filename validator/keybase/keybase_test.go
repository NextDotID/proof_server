package keybase

import (
	"testing"

	"github.com/nextdotid/proof-server/config"
	"github.com/nextdotid/proof-server/types"
	mycrypto "github.com/nextdotid/proof-server/util/crypto"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var (
	kb = Keybase{
		Previous: "",
		Action:   "create",
		Pubkey:   nil,
		Identity: "nykma",
		Platform: types.Platforms.Keybase,
	}
)

func before_each(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	config.Init("../../config/config.test.json")
	pubkey, err := mycrypto.StringToPubkey("0x033d2c5c16bc24ced47619bd3471cef57c8ea8ecce9268700286d61de0d9f3f2dd")
	if err != nil {
		panic(err)
	}
	kb.Pubkey = pubkey
}

func Test_GeneratePostPayload(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)

		new_kb := kb
		result := new_kb.GeneratePostPayload()
		assert.Contains(t, result, "To validate")
		assert.Contains(t, result, mycrypto.CompressedPubkeyHex(new_kb.Pubkey))
		assert.Contains(t, result, "%%SIG_BASE64%%")
		t.Logf("%+v", result)
	})
}

func Test_Validate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)

		new_kb := kb
		new_kb.Identity = "NYKma"
		assert.Nil(t, new_kb.Validate())
		assert.Greater(t, len(new_kb.Signature), 10)
		assert.Equal(t, "nykma", new_kb.Identity)
	})
}
