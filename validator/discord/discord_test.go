package discord

import (
	"testing"

	"github.com/google/uuid"
	"github.com/nextdotid/proof_server/config"
	"github.com/nextdotid/proof_server/types"
	"github.com/nextdotid/proof_server/util"
	"github.com/nextdotid/proof_server/util/crypto"
	"github.com/nextdotid/proof_server/validator"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func before_each() {
	logrus.SetLevel(logrus.DebugLevel)
	config.Init("../../config/config.test.json")
}

func generate() Discord {
	pubkey, _ := crypto.StringToPubkey("0x02d7c5e01bedf1c993f40ec302d9bf162620daea93a7155cd9a8019ae3a2c2a476")
	created_at, _ := util.TimestampStringToTime("1649299881")
	return Discord{
		Base: &validator.Base{
			Platform:      types.Platforms.Discord,
			Previous:      "",
			Action:        types.Actions.Create,
			Pubkey:        pubkey,
			Identity:      "Sannie#0250",
			ProofLocation: "https://discord.com/channels/960708146706395176/960708146706395179/961458176719487076",
			CreatedAt:     created_at,
			Uuid:          uuid.MustParse("27b82012-bc83-4527-9351-9114e500d352"),
		},
	}
}

func TestDiscord_Validate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each()
		discord := generate()
		err := discord.Validate()
		assert.Nil(t, err)
		t.Logf("AltName: %s", discord.AltName)
	})
	t.Run("different user", func(t *testing.T) {
		before_each()
		discord := generate()
		discord.Identity = "test#1234"
		err := discord.Validate()
		assert.NotNil(t, err)
	})
}

func TestDiscord_GenerateSignPayload(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each()
		discord := generate()
		signPayload := discord.GenerateSignPayload()
		assert.Contains(t, signPayload, discord.Identity)
		assert.Contains(t, signPayload, string(discord.Action))
		assert.Contains(t, signPayload, string(discord.Platform))
	})
}

func TestDiscord_GeneratePostPayload(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each()
		discord := generate()
		result := discord.GeneratePostPayload()
		assert.Contains(t, result["default"], discord.Identity)
		assert.Contains(t, result["default"], "%SIG_BASE64%")
	})
}
