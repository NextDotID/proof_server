package slack

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

func generate() Slack {
	pubkey, _ := crypto.StringToPubkey("0x4ec73e36f64ea6e2aa28c101dcae56203e02bd56b4b08c7848b5e791c7bfb9ca2b30f657bd822756533731e201faf57a0aaf6af36bd51f921f7132c9830c6fdf")
	created_at, _ := util.TimestampStringToTime("1677339048")
	uuid := uuid.MustParse("5032b8b3-d91d-434e-be3f-f172267e4006")

	return Slack{
		Base: &validator.Base{
			Platform:      types.Platforms.Slack,
			Previous:      "",
			Action:        types.Actions.Create,
			Pubkey:        pubkey,
			Identity:      "ashfaqur",
			ProofLocation: "https://ashfaqur.slack.com/archives/C04Q3P6H7TK/p1677499644698189",
			CreatedAt:     created_at,
			Uuid:          uuid,
		},
	}
}

func Test_GeneratePostPayload(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)

		slack := generate()
		post := slack.GeneratePostPayload()
		post_default, ok := post["default"]
		require.True(t, ok)
		require.Contains(t, post_default, "Verifying my Slack ID")
		require.Contains(t, post_default, slack.Identity)
		require.Contains(t, post_default, slack.Uuid.String())
		require.Contains(t, post_default, "%SIG_BASE64%")
	})
}

func Test_GenerateSignPayload(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)

		slack := generate()
		payload := slack.GenerateSignPayload()
		require.Contains(t, payload, slack.Uuid.String())
		require.Contains(t, payload, strconv.FormatInt(slack.CreatedAt.Unix(), 10))
		require.Contains(t, payload, slack.Identity)
	})
}

func Test_Validate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)

		slack := generate()
		require.NoError(t, slack.Validate())
		require.Equal(t, "U04Q3NRDWHX", slack.AltID)
	})
}
