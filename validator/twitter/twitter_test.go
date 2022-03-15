package twitter

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

func generate() Twitter {
	pubkey, _ := mycrypto.StringToPubkey("0x037b721d6d84b474edbdab4d0746e9c777f60c414f9b0e651dd08272cb30ed6232")
	created_at, _ := util.TimestampStringToTime("1647327932")
	return Twitter{
		Base: &validator.Base{
			Platform:      types.Platforms.Twitter,
			Previous:      "",
			Action:        types.Actions.Create,
			Pubkey:        pubkey,
			Identity:      "yeiwb",
			ProofLocation: "1503630530465599488",
			Text:          "",
			Uuid:          uuid.MustParse("ed9f421d-92e1-4c80-9bff-8516ef46ff43"),
			CreatedAt:     created_at,
		},
	}
}

func Test_GeneratePostPayload(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)

		tweet := generate()
		result := tweet.GeneratePostPayload()
		assert.Contains(t, result, "Prove myself")
		assert.Contains(t, result, mycrypto.CompressedPubkeyHex(tweet.Pubkey))
		assert.Contains(t, result, "%SIG_BASE64%")
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
