package discord

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/nextdotid/proof-server/config"
	"github.com/nextdotid/proof-server/util"
	"golang.org/x/xerrors"
	"net/url"
	"path"
	"regexp"
	"strings"

	"github.com/nextdotid/proof-server/types"
	"github.com/nextdotid/proof-server/util/crypto"
	"github.com/nextdotid/proof-server/validator"
)

type Discord struct {
	*validator.Base
}

var (
	re            = regexp.MustCompile(MATCH_TEMPLATE)
	POST_TEMPLATE = map[string]string{
		"default": "Verifying my discord ID: %s on NextID. \nSig: %%SIG_BASE64%%",
		"en-US":   "Verifying my discord ID: %s on NextID. \nSig: %%SIG_BASE64%%",
		"zh-CN":   "在NextID上认证我的账号:%s。\nSig: %%SIG_BASE64%%",
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
		"identity":   strings.ToLower(dc.Identity),
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

	u, err := url.Parse(dc.ProofLocation)
	urlPath := path.Clean(u.Path)
	pathArr := strings.Split(strings.TrimSpace(urlPath), "/")
	//proof location will be like: https://discord.com/channels/960708146706395176/960708146706395179/961458176719487076
	if len(pathArr) != 5 {
		return xerrors.Errorf("haven't got right proof location, err=%v", err)
	}

	dg, err := discordgo.New("Bot " + config.C.Platform.Discord.BotToken)
	if err != nil {
		return xerrors.Errorf("error creating Discord session, err= %v", err)
	}
	rs, err := dg.ChannelMessage(pathArr[3], pathArr[4])
	if err != nil {
		return xerrors.Errorf("cannot get the proof err=%v", err)
	}
	if strings.ToLower(fmt.Sprintf("%s", rs.Author)) != dc.Identity {
		return xerrors.Errorf("User name mismatch: expect %s - actual %s", dc.Identity, rs.Author)
	}

	dc.Text = rs.Content
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
		sigBytes, err := base64.StdEncoding.DecodeString(sigBase64)
		if err != nil {
			return xerrors.Errorf("Error when decoding signature %s: %s", sigBase64, err.Error())
		}
		dc.Signature = sigBytes
		return crypto.ValidatePersonalSignature(dc.SignaturePayload, sigBytes, dc.Pubkey)
	}
	return xerrors.Errorf("Signature not found in the message link.")
}
