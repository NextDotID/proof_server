package das

import (
	"testing"

	"github.com/google/uuid"
	"github.com/nextdotid/proof-server/config"
	"github.com/nextdotid/proof-server/types"
	"github.com/nextdotid/proof-server/util"
	mycrypto "github.com/nextdotid/proof-server/util/crypto"
	"github.com/nextdotid/proof-server/validator"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func before_each(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	config.Init("../../config/config.test.json")
}

func generate() Das {
	pubkey, _ := mycrypto.StringToPubkey("0x02fb41da2bf18b9a32afbe9f5163a4390e2de93c421ceb461658b1f83775cdcbc5")
	created_at, _ := util.TimestampStringToTime("1653305520")

	return Das{
		Base: &validator.Base{
			Previous:  "",
			Action:    "create",
			Pubkey:    pubkey,
			Identity:  "nykma.bit",
			Platform:  types.Platforms.Das,
			CreatedAt: created_at,
			Uuid:      uuid.MustParse("77261013-04a6-4464-9d4c-6549114a07a8"),
		},
	}
}

func Test_GeneratePostPayload(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)

		das := generate()
		result := das.GeneratePostPayload()
		assert.Contains(t, result["default"], "To validate")
		assert.Contains(t, result["default"], mycrypto.CompressedPubkeyHex(das.Pubkey))
		assert.Contains(t, result["default"], "%SIG_BASE64%")
	})
}

func Test_Validate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)

		das := generate()
		das.Identity = "NyKmA.BiT"
		assert.Nil(t, das.Validate())
		assert.Greater(t, len(das.Signature), 10)
		assert.Equal(t, "nykma.bit", das.Identity)
	})

	// Do not test validation by ProofLocation, since it is unnecessary.
}
