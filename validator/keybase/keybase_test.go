package keybase

import (
	"testing"

	"github.com/nextdotid/proof-server/config"
	mycrypto "github.com/nextdotid/proof-server/util/crypto"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var (
	kb = Keybase{
		Previous:      "",
		Action:        "create",
		Pubkey:        nil,
		Identity:      "nykma",
	}
)

func before_each(t *testing.T)  {
	logrus.SetLevel(logrus.DebugLevel)
	config.Init("../../config/config.test.json")
	pubkey, err := mycrypto.StringToPubkey("0x033d2c5c16bc24ced47619bd3471cef57c8ea8ecce9268700286d61de0d9f3f2dd")
	if err != nil {
		panic(err)
	}
	kb.Pubkey = pubkey
}

func Test_Validate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)

		newKB := kb
		assert.True(t, newKB.Validate())
		assert.Greater(t, len(newKB.ProofText), 10)
	})
}