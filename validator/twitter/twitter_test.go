package twitter

import (
	"testing"

	"github.com/nextdotid/proof-server/config"
	"github.com/nextdotid/proof-server/types"
	mycrypto "github.com/nextdotid/proof-server/util/crypto"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var (
	tweet = Twitter{
		Previous:      "",
		Action:        types.Actions.Create,
		Pubkey:        nil,
		Identity:      "yeiwb",
		ProofLocation: "1469221200140574721",
	}
)

func before_each(t *testing.T)  {
	logrus.SetLevel(logrus.DebugLevel)
	config.Init("../../config/config.test.json")
	pubkey, err := mycrypto.StringToPubkey("0x028c3cda474361179d653c41a62f6bbb07265d535121e19aedf660da2924d0b1e3")
	if err != nil {
		panic(err)
	}
	tweet.Pubkey = pubkey
	// model.Init()
}

func Test_Validate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)

		newTweet := tweet
		assert.True(t, newTweet.Validate())
		assert.Greater(t, len(newTweet.TweetText), 10)
		assert.Empty(t, tweet.TweetText)
	})

	t.Run("should return identity error", func(t *testing.T) {
		before_each(t)

		newTweet := tweet
		newTweet.Identity = "foobar"
		assert.False(t, newTweet.Validate())
	})

	t.Run("should return proof location not found", func(t *testing.T) {
		before_each(t)

		newTweet := tweet
		newTweet.ProofLocation = "123456"
		assert.False(t, newTweet.Validate())
	})
}
