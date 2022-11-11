package das

import (
	"testing"

	"github.com/google/uuid"
	"github.com/nextdotid/proof_server/config"
	"github.com/nextdotid/proof_server/util"
	mycrypto "github.com/nextdotid/proof_server/util/crypto"
	"github.com/nextdotid/proof_server/validator"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func before_each(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	config.Init("../../config/config.test.json")
}

func generate() Das {
	pubkey, _ := mycrypto.StringToPubkey("0x03b0b5900f2106475027b9f80d249916baa3d0fb57071b9b41980a65868519f825")
	created_at, _ := util.TimestampStringToTime("1653842234")

	return Das{
		Base: &validator.Base{
			Previous:  "",
			Action:    "create",
			Pubkey:    pubkey,
			Identity:  "mitchatmask.bit",
			Platform:  "dotbit",
			CreatedAt: created_at,
			Uuid:      uuid.MustParse("e16a0021-80de-4d12-bea7-9cc021f5b847"),
		},
	}
}

func Test_GeneratePostPayload(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)

		das := generate()
		result := das.GeneratePostPayload()
		require.Contains(t, result["default"], "%SIG_BASE64%")
	})
}

func Test_Validate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)

		das := generate()
		das.Identity = "mItCHaTmASk.BiT"
		require.Nil(t, das.Validate())
		require.Greater(t, len(das.Signature), 10)
		require.Equal(t, "mitchatmask.bit", das.Identity)
		require.Equal(t, das.Identity, das.AltID)
	})

	// Do not test validation by ProofLocation, since it is unnecessary.
}
