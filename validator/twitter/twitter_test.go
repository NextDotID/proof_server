package twitter

import (
	"strings"
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

func generate() Twitter {
	pubkey, _ := mycrypto.StringToPubkey("0x028c3cda474361179d653c41a62f6bbb07265d535121e19aedf660da2924d0b1e3")
	return Twitter{
		Base: &validator.Base{
			Platform:      types.Platforms.Twitter,
			Previous:      "",
			Action:        types.Actions.Create,
			Pubkey:        pubkey,
			Identity:      "YEIwb",
			ProofLocation: "1469221200140574721",
			Text:          "",
		},
	}
}

func Test_GeneratePostPayload(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)

		tweet := generate()
		result := tweet.GeneratePostPayload()
		assert.True(t, strings.Contains(result, "Prove myself"))
		assert.True(t, strings.Contains(result, mycrypto.CompressedPubkeyHex(tweet.Pubkey)))
		assert.True(t, strings.Contains(result, "%SIG_BASE64%"))
	})
}

func Test_Validate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)

		tweet := generate()
		assert.Nil(t, tweet.Validate())
		assert.Greater(t, len(tweet.Text), 10)
		assert.NotEmpty(t, tweet.Text)
		assert.Equal(t, "yeiwb", tweet.Identity)
	})

	t.Run("should return identity error", func(t *testing.T) {
		before_each(t)

		tweet := generate()
		tweet.Identity = "foobar"
		assert.NotNil(t, tweet.Validate())
	})

	t.Run("should return proof location not found", func(t *testing.T) {
		before_each(t)

		tweet := generate()
		tweet.ProofLocation = "123456"
		assert.NotNil(t, tweet.Validate())
	})
}
