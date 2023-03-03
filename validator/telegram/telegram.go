package telegram

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"regexp"
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
	initClient()
	var userId string
	if err := client.Run(context.Background(), func(ctx context.Context) error {
		if _, err := client.Auth().Bot(ctx, config.C.Platform.Telegram.BotToken); err != nil {
			return xerrors.Errorf("Error when authenticating the telegram bot: %v,", err)
		}

		resolved, err := client.API().ContactsResolveUsername(ctx, telegram.Identity)
		if err != nil {
			return xerrors.Errorf("Error while resolving the telegram username: %v,", err)
		}

		if len(resolved.Users) != 1 {
			return xerrors.New("The resulting telegram user is empty")
		}

		user, ok := resolved.Users[0].(*tg.User)
		if !ok {
			return xerrors.New("The resulting telegram user is empty")
		}
		userId = fmt.Sprintf("%d", user.ID)
		return nil
	}); err != nil {
		l.Warnf("Error inside the telegram client context: %v", err)
		return ""
	}

	payloadStruct := validator.H{
		"action":     string(telegram.Action),
		"identity":   userId,
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
	//initClient()
	telegram.Identity = strings.ToLower(telegram.Identity)
	telegram.SignaturePayload = telegram.GenerateSignPayload()

	post, err := validator.GetPostWithHeadlessBrowser(telegram.ProofLocation, telegram.Identity)
	if err != nil {
		return xerrors.Errorf("fetching tweet with headless browser: %w", err)
	}
	telegram.Text = post
	//// Deletion. No need to fetch the telegram message.
	//if telegram.Action == types.Actions.Delete {
	//	return mycrypto.ValidatePersonalSignature(telegram.SignaturePayload, telegram.Signature, telegram.Pubkey)
	//}
	//
	//// Message link of the public channel message, e.g. https://t.me/some_public_channel/CHAT_ID_DIGITS
	//u, err := url.Parse(telegram.ProofLocation)
	//if err != nil {
	//	return xerrors.Errorf("Error when parsing telegram proof location: %v", err)
	//
	//}
	//msgPath := strings.Trim(u.Path, "/")
	//parts := strings.Split(msgPath, "/")
	//if len(parts) != 2 {
	//	return xerrors.Errorf("Error: malformatted telegram proof location: %v", telegram.ProofLocation)
	//}
	//channelName := parts[0]
	//messageId, err := strconv.ParseInt(parts[1], 10, 64)
	//if err != nil {
	//	return xerrors.Errorf("Error when parsing telegram message ID %s: %s", telegram.ProofLocation, err.Error())
	//}
	//
	//// Optional, could be removed
	//if channelName != config.C.Platform.Telegram.PublicChannelName {
	//	return xerrors.New("Unknown channel")
	//}
	//
	//if err := client.Run(context.Background(), func(ctx context.Context) error {
	//
	//	if _, err := client.Auth().Bot(ctx, config.C.Platform.Telegram.BotToken); err != nil {
	//		return xerrors.Errorf("Error when authenticating the telegram bot: %v,", err)
	//	}
	//
	//	resolved, err := client.API().ContactsResolveUsername(ctx, channelName)
	//	if err != nil {
	//		return xerrors.Errorf("Error while resolving the public channel name: %v,", err)
	//	}
	//
	//	if len(resolved.Chats) != 1 {
	//		return xerrors.New("The resulting telegram public channel is empty")
	//	}
	//
	//	channel, ok := resolved.Chats[0].(*tg.Channel)
	//	if !ok {
	//		return xerrors.New("The resulting telegram public channel is empty")
	//	}
	//
	//	msgsClass, err := client.API().ChannelsGetMessages(ctx, &tg.ChannelsGetMessagesRequest{
	//		Channel: &tg.InputChannel{
	//			ChannelID:  channel.ID,
	//			AccessHash: channel.AccessHash,
	//		},
	//		ID: []tg.InputMessageClass{
	//			&tg.InputMessageID{ID: int(messageId)},
	//		},
	//	})
	//
	//	if err != nil {
	//		return xerrors.Errorf("Error while fetching the public channel message: %v,", err)
	//	}
	//
	//	msgList, ok := msgsClass.(*tg.MessagesChannelMessages)
	//	if !ok || len(msgList.Messages) != 1 || len(msgList.Messages) != 2 {
	//		return xerrors.New("Please try again sending an original message")
	//	}
	//	user, userOk := msgList.Users[0].(*tg.User)
	//	if !userOk {
	//		return xerrors.New("Please try again sending an original message")
	//	}
	//	if user.Bot {
	//		user, userOk = msgList.Users[1].(*tg.User)
	//	}
	//	msg, msgOk := msgList.Messages[0].(*tg.Message)
	//	if !msgOk || !userOk {
	//		return xerrors.New("Please try again sending an original message")
	//	}
	//	userId := strconv.FormatInt(user.ID, 10)
	//	if strings.EqualFold(userId, telegram.Identity) {
	//		return xerrors.Errorf("Telegram username mismatch: expect %s - actual %s", telegram.Identity, user.Username)
	//	}
	//
	//	telegram.Text = msg.Message
	//	telegram.AltID = user.Username
	//	telegram.Identity = userId
	//	return telegram.validateText()
	//}); err != nil {
	//	return xerrors.Errorf("Error inside the telegram client context: %v", err)
	//}
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

	// Exit if the we can't authenticate the telegram client
	// using the provided configs.
	if err := client.Run(context.Background(), func(ctx context.Context) error {
		if _, err := client.Auth().Bot(ctx, config.C.Platform.Telegram.BotToken); err != nil {
			return xerrors.Errorf("Error when authenticating the telegram bot: %v,", err)
		}
		return nil
	}); err != nil {
		panic(err)
	}

}
