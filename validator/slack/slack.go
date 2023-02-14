package slack

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"string"

	"github.com/nextdotid/proof_server/config"
	types "github.com/nextdotid/proof_server/types"
	util "github.com/nextdotid/proof_server/util"
	mycrypto "github.com/nextdotid/proof_server/util/crypto"
	"github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	slackClient "github.com/slack-go/slack"
	"golang.org/x/xerrors"

	validator "github.com/nextdotid/proof_server/validator"
)

// Slack represents the validator for Slack platform
type Slack struct {
	*validator.Base
}

const (
	matchTemplate = "^Sig: (.*)$"
)

var (
	client     *slackClient.Client
	l          = logrus.WithFields(logrus.Fields{"module": "validator", "validator": "slack"})
	re         = regexp.MustCompile(matchTemplate)
	postStruct = map[string]string{
		"default": "🎭 Verifying my Slack ID @%s for @NextDotID.\nSig: %%SIG_BASE64%%\n\nPowered by Next.ID - Connect All Digital Identities.\n",
		"en_US":   "🎭 Verifying my Slack ID @%s for @NextDotID.\nSig: %%SIG_BASE64%%\n\nPowered by Next.ID - Connect All Digital Identities.\n",
		"zh_CN":   "🎭 正在通过 @NextDotID 验证我的 Slack 帐号 @%s 。\nSig: %%SIG_BASE64%%\n\n由 Next.ID 支持 - 连接全域数字身份。\n",
	}
)

// Init initializes the Slack validator
func Init() {
	initClient()
	if validator.PlatformFactories == nil {
		validator.PlatformFactories = make(map[types.Platform]func(*validator.Base) validator.IValidator)
	}
	validator.PlatformFactories[types.Platforms.Slack] = func(base *validator.Base) validator.IValidator {
		slack := &Slack{base}
		return slack
	}
}

// GeneratePostPayload generates the post payload for Slack
func (s *Slack) GeneratePostPayload() (post map[string]string) {
	post = make(map[string]string)
	for langCode, template := range postStruct {
		post[langCode] = fmt.Sprintf(template, s.Identity)
	}
	return post
}

// GenerateSignPayload generates the signature payload for Slack
func (slack *Slack) GenerateSignPayload() (payload string) {
	payloadStruct := validator.H{
		"action":     string(slack.Action),
		"identity":   slack.Identity,
		"platform":   "slack",
		"created_at": util.TimeToTimestampString(slack.CreatedAt),
		"uuid":       slack.Uuid.String(),
	}
	if slack.Previous != "" {
		payloadStruct["prev"] = slack.Previous
	}

	payloadBytes, err := json.Marshal(payloadStruct)
	if err != nil {
		l.Warnf("Error when marshaling struct: %s", err.Error())
		return ""
	}

	return string(payloadBytes)
}

func (slack *Slack) Validate() (err error) {
	client := initClient()
	slack.Identity = strings.ToLower(slack.Identity)
	slack.SignaturePayload = slack.GenerateSignPayload()

	if slack.Action == types.Actions.Delete {
		return mycrypto.ValidatePersonalSignature(slack.SignaturePayload, slack.Signature, slack.Pubkey)
	}

	u, err := url.Parse(slack.ProofLocation)
	if err != nil {
		return xerrors.Errorf("Error when parsing slack proof location: %v", err)
	}
	msgPath := strings.Trim(u.Path, "/")
	parts := strings.Split(msgPath, "/")
	if len(parts) != 2 {
		return xerrors.Errorf("Error: malformatted slack proof location: %v", slack.ProofLocation)
	}
	channelID := parts[0]
	messageID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return xerrors.Errorf("Error when parsing slack message ID %s: %s", slack.ProofLocation, err.Error())
	}

	msgResp, err := client.GetChannelMessage(channelID, messageID)
	if err != nil {
		return xerrors.Errorf("Error getting the message from slack: %w", err)
	}

    user := msgResp.User
	userID := strconv.FormatInt(user.ID, 10)
	if !strings.EqualFold(userID, slack.Identity) {
		return xerrors.Errorf("slack userID mismatch: expect %s - actual %s", slack.Identity, userID)
	}

	slack.Text = msgResp.Message
	slack.AltID = userID
	slack.Identity = userID

	return slack.validateText()
}

func (slack *Slack) validateText() (err error) {
	scanner := bufio.NewScanner(strings.NewReader(slack.Text))
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
		slack.Signature = sigBytes
		return mycrypto.ValidatePersonalSignature(slack.SignaturePayload, sigBytes, slack.Pubkey)
	}
	return xerrors.Errorf("Signature not found in the slack message.")
}

var client *slack.Client

func initClient() *slack.Client {
	if client == nil {
        httpClient := httpClient{}
	    client = slack.New(config.C.Platform.Slack.ApiToken, slack.OptionHTTPClient(&httpClient))
	}
	return client			

}
