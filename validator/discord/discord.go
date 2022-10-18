package discord

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/nextdotid/proof_server/config"
	"github.com/nextdotid/proof_server/util"
	"golang.org/x/xerrors"

	"github.com/nextdotid/proof_server/types"
	"github.com/nextdotid/proof_server/util/crypto"
	"github.com/nextdotid/proof_server/validator"
)

type Discord struct {
	*validator.Base
}

var (
	re            = regexp.MustCompile(MATCH_TEMPLATE)
	POST_TEMPLATE = map[string]string{
		"default": "Verifying my discord ID: %s on NextID. \nSig: %%SIG_BASE64%%",
		"en-US":   "Verifying my discord ID: %s on NextID. \nSig: %%SIG_BASE64%%",
		"zh-CN":   "在NextID上认证我的账号： %s \nSig: %%SIG_BASE64%%",
	}
)

const (
	MATCH_TEMPLATE = "^Sig: (.*)$"
)

func Init() {
	if validator.PlatformFactories == nil {
		validator.PlatformFactories = make(map[types.Platform]func(*validator.Base) validator.IValidator)
	}
	validator.PlatformFactories[types.Platforms.Discord] = func(base *validator.Base) validator.IValidator {
		dc := Discord{base}
		return &dc
	}
}

func (dc *Discord) GeneratePostPayload() (post map[string]string) {
	post = make(map[string]string, 0)
	for lang_code, template := range POST_TEMPLATE {
		post[lang_code] = fmt.Sprintf(template, dc.Identity)
	}
	return post
}

func (dc *Discord) GenerateSignPayload() (payload string) {
	payloadStruct := validator.H{
		"action":     string(dc.Action),
		"identity":   dc.Identity,
		"platform":   string(types.Platforms.Discord),
		"prev":       nil,
		"created_at": util.TimeToTimestampString(dc.CreatedAt),
		"uuid":       dc.Uuid.String(),
	}

	if dc.Previous != "" {
		payloadStruct["prev"] = dc.Previous
	}

	payload_bytes, _ := json.Marshal(payloadStruct)
	return string(payload_bytes)
}

func (dc *Discord) Validate() (err error) {
	dc.SignaturePayload = dc.GenerateSignPayload()

	// Delete. No need to fetch content from platform.
	if dc.Action == types.Actions.Delete {
		return crypto.ValidatePersonalSignature(dc.SignaturePayload, dc.Signature, dc.Pubkey)
	}

	u, err := url.Parse(dc.ProofLocation)
	urlPath := path.Clean(u.Path)
	pathArr := strings.Split(strings.TrimSpace(urlPath), "/")

	//proof location will be like: https://discord.com/channels/960708146706395176/960708146706395179/961458176719487076
	if len(pathArr) != 5 {
		return xerrors.Errorf("Error getting right proof location: %w", err)
	}

	client, err := discordgo.New("Bot " + config.C.Platform.Discord.BotToken)
	if err != nil {
		return xerrors.Errorf("Error creating Discord session: %w", err)
	}

	msgResp, err := client.ChannelMessage(pathArr[3], pathArr[4])
	if err != nil {
		return xerrors.Errorf("Error getting the message from discord: %w", err)
	}

	if fmt.Sprintf("%s", msgResp.Author) != dc.Identity {
		return xerrors.Errorf("User name mismatch: expect %s - actual %s", dc.Identity, msgResp.Author)
	}

	dc.AltID = msgResp.Author.ID
	dc.Text = msgResp.Content
	return dc.validateText()
}

func (dc *Discord) validateText() (err error) {
	scanner := bufio.NewScanner(strings.NewReader(dc.Text))
	for scanner.Scan() {
		matched := re.FindStringSubmatch(scanner.Text())
		if len(matched) < 2 {
			continue // Search for next line
		}
		sigBase64 := matched[1]
		sigBytes, err := util.DecodeString(sigBase64)
		if err != nil {
			return xerrors.Errorf("Error when decoding signature %s: %s", sigBase64, err.Error())
		}
		dc.Signature = sigBytes
		return crypto.ValidatePersonalSignature(dc.SignaturePayload, sigBytes, dc.Pubkey)
	}
	return xerrors.Errorf("Signature not found in the message link.")
}
