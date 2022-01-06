package discord

import (
	"encoding/json"
	"fmt"

	"github.com/nextdotid/proof-server/config"
	"github.com/nextdotid/proof-server/types"
	"github.com/nextdotid/proof-server/util/crypto"
	"github.com/nextdotid/proof-server/validator"

	"github.com/bwmarrin/discordgo"
)

// Discord.Identity: Discord User ID (digits, not Name#1234)
type Discord validator.Base

const (
	POST_TEMPLATE = "Prove myself: I'm 0x%s on NextID. Signature: %%SIG_BASE64%%"
)

func Init() {
	if validator.PlatformFactories == nil {
		validator.PlatformFactories = make(map[types.Platform]func(validator.Base) validator.IValidator)
	}
	validator.PlatformFactories[types.Platforms.Discord] = func(base validator.Base) validator.IValidator {
		dis := Discord(base)
		return &dis
	}

	discord, err := discordgo.New("Bot " + config.C.Platform.Discord.BotToken)
	if err != nil {
		panic(fmt.Sprintf("Discord init err: %s", err))
	}


	discord.ChannelMessage(config.C.Platform.Discord.ProofServerChannelID, "")
}

func (dis *Discord) GeneratePostPayload() (payload string) {
	persona := crypto.CompressedPubkeyHex(dis.Pubkey)
	return fmt.Sprintf(POST_TEMPLATE, persona)
}

func (dis *Discord) GenerateSignPayload() (payload string) {
	payloadStruct := validator.H{
		"action": string(dis.Action),
		"identity": dis.Identity,
		"platform": string(types.Platforms.Discord),
		"prev": nil,
	}

	if dis.Previous != "" {
		payloadStruct["prev"] = dis.Previous
	}

	payload_bytes, _ := json.Marshal(payloadStruct)
	return string(payload_bytes)
}

func (dis *Discord) Validate() (err error) {
	return nil // TODO
}
