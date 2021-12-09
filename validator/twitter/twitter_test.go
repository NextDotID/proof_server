package twitter

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/nextdotid/proof-server/config"
	"github.com/nextdotid/proof-server/types"
	"github.com/stretchr/testify/assert"
)

var (
	tweet = Twitter{
		Previous:      "",
		Action:        types.Actions.Create,
		Pubkey:        common.HexToHash("0x1234"),
		Identity:      "846kizuQ",
		ProofLocation: "1466752395921477633",
	}
)

func before_each(t *testing.T)  {
	config.Init("../../config/config.test.json")
	// model.Init()
}

func Test_Validate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)

		newTweet := tweet
		assert.False(t, newTweet.Validate())
		assert.Greater(t, len(newTweet.TweetText), 10)
		assert.Empty(t, tweet.TweetText)
	})

	t.Run("should return identity error", func(t *testing.T) {
		before_each(t)

		newTweet := tweet
		newTweet.Identity = "foobar"
		assert.False(t, tweet.Validate())
	})

	t.Run("should return proof location not found", func(t *testing.T) {
		before_each(t)

		newTweet := tweet
		newTweet.ProofLocation = "123456"
		assert.False(t, newTweet.Validate())
	})
}
