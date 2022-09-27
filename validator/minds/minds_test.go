package minds

import (
	"strconv"
	"testing"

	"github.com/google/uuid"
	"github.com/nextdotid/proof_server/config"
	"github.com/nextdotid/proof_server/types"
	"github.com/nextdotid/proof_server/util"
	"github.com/nextdotid/proof_server/util/crypto"
	"github.com/nextdotid/proof_server/validator"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func before_each(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	config.Init("../../config/config.test.json")
}

func generate() Minds {
	pubkey, _ := crypto.StringToPubkey("0x0398a22485635ed2262094103cfdc1511b785011e32a2eb16e5b32fd8561ea6ad8")
	created_at, _ := util.TimestampStringToTime("1664179121")
	uuid := uuid.MustParse("3d770975-5085-411b-91e4-661bcc407aa9")

	return Minds{
		Base: &validator.Base{
			Platform:      types.Platforms.Minds,
			Previous:      "",
			Action:        types.Actions.Create,
			Pubkey:        pubkey,
			Identity:      "nykma",
			ProofLocation: "1421043369127186449",
			CreatedAt:     created_at,
			Uuid:          uuid,
		},
	}
}

func Test_GeneratePostPayload(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)

		minds := generate()
		post := minds.GeneratePostPayload()
		post_default, ok := post["default"]
		require.True(t, ok)
		require.Contains(t, post_default, "Verifying my Minds ID")
		require.Contains(t, post_default, minds.Identity)
		require.Contains(t, post_default, minds.Uuid.String())
		require.Contains(t, post_default, "%SIG_BASE64%")
	})
}

func Test_GenerateSignPayload(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)

		minds := generate()
		payload := minds.GenerateSignPayload()
		require.Contains(t, payload, minds.Uuid.String())
		require.Contains(t, payload, strconv.FormatInt(minds.CreatedAt.Unix(), 10))
		require.Contains(t, payload, minds.Identity)
	})
}

func Test_Validate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)

		minds := generate()
		require.NoError(t, minds.Validate())
	})
}
