package telegram

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/nextdotid/proof_server/config"
	"github.com/nextdotid/proof_server/types"
	"github.com/nextdotid/proof_server/util"
	mycrypto "github.com/nextdotid/proof_server/util/crypto"
	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"github.com/nextdotid/proof_server/validator"
)

type Telegram struct {
	*validator.Base
}

const (
	MATCH_TEMPLATE = "^Sig: (.*)$"
)

var (
	client      *telegram.Client
	l           = logrus.WithFields(logrus.Fields{"module": "validator", "validator": "telegram"})
	re          = regexp.MustCompile(MATCH_TEMPLATE)
	POST_STRUCT = map[string]string{
		"default": "🎭 Verifying my Telegram ID @%s for @NextDotID.\nSig: %%SIG_BASE64%%\n\nPowered by Next.ID - Connect All Digital Identities.\n",
		"en_US":   "🎭 Verifying my Telegram ID @%s for @NextDotID.\nSig: %%SIG_BASE64%%\n\nPowered by Next.ID - Connect All Digital Identities.\n",
		"zh_CN":   "🎭 正在通过 @NextDotID 验证我的 Telegram 帐号 @%s 。\nSig: %%SIG_BASE64%%\n\n由 Next.ID 支持 - 连接全域数字身份。\n",
	}
)

func Init() {
	initClient()
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
	telegram.Identity = strings.ToLower(telegram.Identity)
	payloadStruct := validator.H{
		"action":     string(telegram.Action),
		"identity":   telegram.Identity,
		"platform":   "telegram",
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
	initClient()
	telegram.Identity = strings.ToLower(telegram.Identity)
	telegram.SignaturePayload = telegram.GenerateSignPayload()
	// Deletion. No need to fetch the telegram message.
	if telegram.Action == types.Actions.Delete {
		return mycrypto.ValidatePersonalSignature(telegram.SignaturePayload, telegram.Signature, telegram.Pubkey)
	}

	// Id of the direct message has been sent to the bot.
	messageId, err := strconv.ParseInt(telegram.ProofLocation, 10, 64)
	if err != nil {
		return xerrors.Errorf("Error when parsing telegram message ID %s: %s", telegram.ProofLocation, err.Error())
	}

	if err := client.Run(context.Background(), func(ctx context.Context) error {

		if _, err := client.Auth().Bot(ctx, config.C.Platform.Telegram.BotToken); err != nil {
			return xerrors.Errorf("Error when authenticating the telegram bot: %v,", err)
		}

		msgsClass, err := client.API().MessagesGetMessages(ctx, []tg.InputMessageClass{
			&tg.InputMessageID{
				ID: int(messageId),
			},
		})

		if err != nil {
			return xerrors.Errorf("Error while fetching the user direct message (should be sent to the bot directly): %v,", err)
		}

		msgList, ok := msgsClass.(*tg.MessagesMessages)
		if !ok || len(msgList.Messages) != 1 || len(msgList.Messages) != 2 {
			return xerrors.New("Please try again sending an original message 1")
		}
		user, userOk := msgList.Users[0].(*tg.User)
		if !userOk {
			return xerrors.New("Please try again sending an original message 2")
		}
		if user.Bot {
			user, userOk = msgList.Users[1].(*tg.User)

		}
		msg, msgOk := msgList.Messages[0].(*tg.Message)
		if !msgOk || !userOk {
			return xerrors.New("Please try again sending an original message 2")
		}

		if strings.EqualFold(user.Username, telegram.Identity) {
			return xerrors.Errorf("Screen name mismatch: expect %s - actual %s", telegram.Identity, user.Username)
		}
		telegram.Text = msg.Message
		telegram.AltID = strconv.FormatInt(user.ID, 10)
		return telegram.validateText()
	}); err != nil {
		return xerrors.Errorf("Error inside the telegram client context: %v", err)
	}
	return xerrors.New("Unknown error")
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

func initClient() {
	if client != nil {
		return
	}
	// https://core.telegram.org/api/obtaining_api_id
	client = telegram.NewClient(config.C.Platform.Telegram.ApiID, config.C.Platform.Telegram.ApiHash, telegram.Options{})

}
