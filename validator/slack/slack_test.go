package slack

import (
	"testing"

	"github.com/google/uuid"
	"github.com/nextdotid/proof_server/config"
	"github.com/nextdotid/proof_server/types"
	"github.com/nextdotid/proof_server/util"
	mycrypto "github.com/nextdotid/proof_server/util/crypto"
	"github.com/nextdotid/proof_server/validator"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func (s *Slack) GeneratePostPayload() map[string]string {
	payload := make(map[string]string)
	payload["default"] = "Verifying my Slack ID: " + s.Identity + "\n\n" + s.Text + "\n\n" + "%SIG_BASE64%"
	return payload
}

func before_each(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	config.Init("../../config/config.test.json")
}

func generate() Slack {
	pubkey, _ := mycrypto.StringToPubkey("0x04666b700aeb6a6429f13cbb263e1bc566cd975a118b61bc796204109c1b351d19b7df23cc47f004e10fef41df82bad646b027578f8881f5f1d2f70c80dfcd8031")
	created_at, _ := util.TimestampStringToTime("1647503071")
	return Slack{
		Base: &validator.Base{
			Platform:      types.Platforms.Slack,
			Previous:      "",
			Action:        types.Actions.Create,
			Pubkey:        pubkey,
			Identity:      "yeiwb",
			ProofLocation: "https://app.slack.com/client/T04PR0R7DC3/C04Q3P6H7TK",
			Text:          "",
			Uuid:          uuid.MustParse("c6fa1483-1bad-4f07-b661-678b191ab4b3"),
			CreatedAt:     created_at,
		}
	}
}

func generate() Slack {
	pubkey, _ := mycrypto.StringToPubkey("0x04d7c5e01bedf1c993f40ec302d9bf162620daea93a7155cd9a8019ae3a2c2a476873e66c7ab9c5dbf9a6bd24ef4432298e70c5c7e7b148a54724a1d7b59e06bd8")
	created_at, _ := util.TimestampStringToTime("1650883741")
	return Slack{
		Base: &validator.Base{
			Platform:      types.Platforms.Slack,
			Previous:      "",
			Action:        types.Actions.Create,
			Pubkey:        pubkey,
			Identity:      "SannieInMeta",
			ProofLocation: "https://app.slack.com/client/T04PR0R7DC3/C04Q3P6H7TK",
			Text:          "",
			Uuid:          uuid.MustParse("223a5c86-540b-49b7-8674-94e04a390cd0"),
			CreatedAt:     created_at,
		}
	}
}

func Test_GeneratePostPayload(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)
		slack := generate()
		result := slack.GeneratePostPayload()
		require.Contains(t, result["default"], "Verifying my Slack ID")
		require.Contains(t, result["default"], slack.Identity)
		require.Contains(t, result["default"], "%SIG_BASE64%")
	})
}

func Test_Validate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)

		message := generate()
		require.Nil(t, message.Validate())
		require.Greater(t, len(message.Text), 10)
		require.NotEmpty(t, message.Text)
		require.Equal(t, "yeiwb", message.Identity)
		require.Equal(t, "1468853291941773312", message.AltID)
	})

	t.Run("success on encode base1024", func(t *testing.T) {
		before_each(t)
		message := generateBase1024Encode()
		require.Nil(t, message.Validate())
		require.Greater(t, len(message.Text), 10)
		require.NotEmpty(t, message.Text)
		require.Equal(t, "sannieinmeta", message.Identity)
	})

	t.Run("should return identity error", func(t *testing.T) {
		before_each(t)

		message := generate()
		message.Identity = "foobar"
		require.NotNil(t, message.Validate())
	})

	t.Run("should return proof location not found", func(t *testing.T) {
		before_each(t)
		message := generate()
		message.ProofLocation = "123456"
		require.NotNil(t, message.Validate())
	})
}
