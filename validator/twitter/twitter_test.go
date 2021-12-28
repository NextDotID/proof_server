package twitter

import (
	"strings"
	"testing"

	"github.com/nextdotid/proof-server/config"
	"github.com/nextdotid/proof-server/types"
	mycrypto "github.com/nextdotid/proof-server/util/crypto"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var (
	tweet = Twitter{
		Platform:      types.Platforms.Twitter,
		Previous:      "",
		Action:        types.Actions.Create,
		Pubkey:        nil,
		Identity:      "YEIwb",
		ProofLocation: "1469221200140574721",
		Text:          "",
	}
)

func before_each(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	config.Init("../../config/config.test.json")
	pubkey, err := mycrypto.StringToPubkey("0x028c3cda474361179d653c41a62f6bbb07265d535121e19aedf660da2924d0b1e3")
	if err != nil {
		panic(err)
	}
	tweet.Pubkey = pubkey
}

func Test_GeneratePostPayload(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)

		result := tweet.GeneratePostPayload()
		assert.True(t, strings.Contains(result, "Prove myself"))
		assert.True(t, strings.Contains(result, mycrypto.CompressedPubkeyHex(tweet.Pubkey)))
		assert.True(t, strings.Contains(result, "%SIG_BASE64%"))
	})
}

func Test_Validate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)

		newTweet := tweet
		assert.Nil(t, newTweet.Validate())
		assert.Greater(t, len(newTweet.Text), 10)
		assert.Empty(t, tweet.Text)
		assert.Equal(t, "yeiwb", newTweet.Identity)
	})

	t.Run("should return identity error", func(t *testing.T) {
		before_each(t)

		newTweet := tweet
		newTweet.Identity = "foobar"
		assert.NotNil(t, newTweet.Validate())
	})

	t.Run("should return proof location not found", func(t *testing.T) {
		before_each(t)

		newTweet := tweet
		newTweet.ProofLocation = "123456"
		assert.NotNil(t, newTweet.Validate())
	})
}
