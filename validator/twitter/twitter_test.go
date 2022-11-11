package twitter

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

func before_each(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	config.Init("../../config/config.test.json")
}

func generate() Twitter {
	pubkey, _ := mycrypto.StringToPubkey("0x04666b700aeb6a6429f13cbb263e1bc566cd975a118b61bc796204109c1b351d19b7df23cc47f004e10fef41df82bad646b027578f8881f5f1d2f70c80dfcd8031")
	created_at, _ := util.TimestampStringToTime("1647503071")
	return Twitter{
		Base: &validator.Base{
			Platform:      types.Platforms.Twitter,
			Previous:      "",
			Action:        types.Actions.Create,
			Pubkey:        pubkey,
			Identity:      "yeiwb",
			ProofLocation: "1504363098328924163",
			Text:          "",
			Uuid:          uuid.MustParse("c6fa1483-1bad-4f07-b661-678b191ab4b3"),
			CreatedAt:     created_at,
		},
	}
}

func generateBase1024Encode() Twitter {
	pubkey, _ := mycrypto.StringToPubkey("0x04d7c5e01bedf1c993f40ec302d9bf162620daea93a7155cd9a8019ae3a2c2a476873e66c7ab9c5dbf9a6bd24ef4432298e70c5c7e7b148a54724a1d7b59e06bd8")
	created_at, _ := util.TimestampStringToTime("1650883741")
	return Twitter{
		Base: &validator.Base{
			Platform:      types.Platforms.Twitter,
			Previous:      "",
			Action:        types.Actions.Create,
			Pubkey:        pubkey,
			Identity:      "SannieInMeta",
			ProofLocation: "1518542666987819009",
			Text:          "",
			Uuid:          uuid.MustParse("223a5c86-540b-49b7-8674-94e04a390cd0"),
			CreatedAt:     created_at,
		},
	}
}

func Test_GeneratePostPayload(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)
		tweet := generate()
		result := tweet.GeneratePostPayload()
		require.Contains(t, result["default"], "Verifying my Twitter ID")
		require.Contains(t, result["default"], tweet.Identity)
		require.Contains(t, result["default"], "%SIG_BASE64%")
	})
}

func Test_Validate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)

		tweet := generate()
		require.Nil(t, tweet.Validate())
		require.Greater(t, len(tweet.Text), 10)
		require.NotEmpty(t, tweet.Text)
		require.Equal(t, "yeiwb", tweet.Identity)
		require.Equal(t, "1468853291941773312", tweet.AltID)
	})

	t.Run("success on encode base1024", func(t *testing.T) {
		before_each(t)
		tweet := generateBase1024Encode()
		require.Nil(t, tweet.Validate())
		require.Greater(t, len(tweet.Text), 10)
		require.NotEmpty(t, tweet.Text)
		require.Equal(t, "sannieinmeta", tweet.Identity)
	})

	t.Run("should return identity error", func(t *testing.T) {
		before_each(t)

		tweet := generate()
		tweet.Identity = "foobar"
		require.NotNil(t, tweet.Validate())
	})

	t.Run("should return proof location not found", func(t *testing.T) {
		before_each(t)
		tweet := generate()
		tweet.ProofLocation = "123456"
		require.NotNil(t, tweet.Validate())
	})
}
