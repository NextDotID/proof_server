package telegram

import (
	"bufio"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/nextdotid/proof_server/types"
	"github.com/nextdotid/proof_server/util"
	mycrypto "github.com/nextdotid/proof_server/util/crypto"
	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"

	"github.com/nextdotid/proof_server/validator"
)

type Telegram struct {
	*validator.Base
}

const (
	MATCH_TEMPLATE = "^Sig: (.*)$"
)

var (
	l           = logrus.WithFields(logrus.Fields{"module": "validator", "validator": "telegram"})
	re          = regexp.MustCompile(MATCH_TEMPLATE)
	POST_STRUCT = map[string]string{
		"default": "🎭 Verifying my Telegram ID @%s for @NextDotID.\nSig: %%SIG_BASE64%%\n\nPowered by Next.ID - Connect All Digital Identities.\n",
		"en_US":   "🎭 Verifying my Telegram ID @%s for @NextDotID.\nSig: %%SIG_BASE64%%\n\nPowered by Next.ID - Connect All Digital Identities.\n",
		"zh_CN":   "🎭 正在通过 @NextDotID 验证我的 Telegram 帐号 @%s 。\nSig: %%SIG_BASE64%%\n\n由 Next.ID 支持 - 连接全域数字身份。\n",
	}
)

func Init() {
	if validator.PlatformFactories == nil {
		validator.PlatformFactories = make(map[types.Platform]func(*validator.Base) validator.IValidator)
	}

	validator.PlatformFactories[types.Platforms.Telegram] = func(base *validator.Base) validator.IValidator {
		telg := Telegram{base}
		return &telg
	}
}

func (telegram *Telegram) GeneratePostPayload() (post map[string]string) {
	post = make(map[string]string, 0)
	for lang_code, template := range POST_STRUCT {
		post[lang_code] = fmt.Sprintf(template, telegram.Identity)
	}

	return post
}

func (telegram *Telegram) GenerateSignPayload() (payload string) {
	payloadStruct := validator.H{
		"action":     string(telegram.Action),
		"identity":   telegram.Identity,
		"platform":   string(types.Platforms.Telegram),
		"prev":       nil,
		"created_at": util.TimeToTimestampString(telegram.CreatedAt),
		"uuid":       telegram.Uuid.String(),
	}
	if telegram.Previous != "" {
		payloadStruct["prev"] = telegram.Previous
	}

	payloadBytes, err := json.Marshal(payloadStruct)
	if err != nil {
		l.Warnf("Error when marshaling struct: %s", err.Error())
		return ""
	}

	return string(payloadBytes)
}

func (telegram *Telegram) Validate() (err error) {
	telegram.Identity = strings.ToLower(telegram.Identity)
	telegram.SignaturePayload = telegram.GenerateSignPayload()

	//// Deletion. No need to fetch the telegram message.
	if telegram.Action == types.Actions.Delete {
		return mycrypto.ValidatePersonalSignature(telegram.SignaturePayload, telegram.Signature, telegram.Pubkey)
	}

	userLink, err := validator.GetPostWithHeadlessBrowser(fmt.Sprintf("%s%s", telegram.ProofLocation, "?embed=1&mode=tme"), "div.tgme_widget_message_user", "^$", "href")
	if err != nil {
		return xerrors.Errorf("fetching post message with headless browser: %w", err)
	}

	username := userLink[strings.LastIndex(userLink, "/")+1 : len(userLink)]
	if username != telegram.Identity {
		return xerrors.Errorf("User name mismatch: expect %s - actual %s", telegram.Identity, username)
	}

	post, err := validator.GetPostWithHeadlessBrowser(fmt.Sprintf("%s%s", telegram.ProofLocation, "?embed=1&mode=tme"), "div.tgme_widget_message_text.js-message_text", "Sig:", "text")
	if err != nil {
		return xerrors.Errorf("fetching post message with headless browser: %w", err)
	}

	telegram.Text = post
	return telegram.validateText()
}

func (telegram *Telegram) validateText() (err error) {
	scanner := bufio.NewScanner(strings.NewReader(telegram.Text))
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
		telegram.Signature = sigBytes
		return mycrypto.ValidatePersonalSignature(telegram.SignaturePayload, sigBytes, telegram.Pubkey)
	}
	return xerrors.Errorf("Signature not found in the telegram message.")
}
