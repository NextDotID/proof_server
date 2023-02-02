package slack

import (
    "bufio"
    "encoding/json"
    "fmt"
    "net/url"
    "regexp"
    "strconv"
    "strings"

"github.com/nlopes/slack"

"github.com/nextdotid/proof_server/config"
    "github.com/nextdotid/proof_server/types"
    "github.com/nextdotid/proof_server/util"
    mycrypto "github.com/nextdotid/proof_server/util/crypto"
    "github.com/sirupsen/logrus"
    "golang.org/x/xerrors"

"github.com/nextdotid/proof_server/validator"
)

type Slack struct {
    *validator.Base
}

const (
    MATCH_TEMPLATE = "^Sig: (.*)$"
)

var (
    client      *slack.Client
    l           = logrus.WithFields(logrus.Fields{"module": "validator", "validator": "slack"})
    re          = regexp.MustCompile(MATCH_TEMPLATE)
    POST_STRUCT = map[string]string{
        "default": "🎭 Verifying my Slack ID @%s for @NextDotID.\nSig: %%SIG_BASE64%%\n\nPowered by Next.ID - Connect All Digital Identities.\n",
        "en_US":   "🎭 Verifying my Slack ID @%s for @NextDotID.\nSig: %%SIG_BASE64%%\n\nPowered by Next.ID - Connect All Digital Identities.\n",
        "zh_CN":   "🎭 正在通过 @NextDotID 验证我的 Slack 帐号 @%s 。\nSig: %%SIG_BASE64%%\n\n由 Next.ID 支持 - 连接全域数字身份。\n",
    }
)

func Init() {
    initClient()
    if validator.PlatformFactories == nil {
        validator.PlatformFactories = make(map[types.Platform]func(*validator.Base) validator.IValidator)
    }
    validator.PlatformFactories[types.Platforms.Slack] = func(base *validator.Base) validator.IValidator {
        slack :=Slack{base}
        return &slack
    }
}

func (slack *Slack) GeneratePostPayload() (post map[string]string) {
    post = make(map[string]string, 0)
    for lang_code, template := range POST_STRUCT {
        post[lang_code] = fmt.Sprintf(template, slack.Identity)
    }

    return post
}
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
    initClient()
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
    channelName := parts[0]
    messageId, err := strconv.ParseInt(parts[1], 10, 64)
    if err != nil {
        return xerrors.Errorf("Error when parsing slack message ID %s: %s", slack.ProofLocation, err.Error())
    }

    user, err := GetUser(channelName, messageId)
    if err != nil {
        return xerrors.Errorf("Error when fetching user from slack: %v", err)
    }

    userId := strconv.FormatInt(user.ID, 10)
    if strings.EqualFold(userId, slack.Identity) {
        return xerrors.Errorf("slack username mismatch: expect %s - actual %s", slack.Identity, user.Username)
    }

    slack.Text = msg.Message
    slack.AltID = user.Username
    slack.Identity = userId

    return slack.ValidateText()
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
            return errors.Wrapf(err, "Error when decoding signature %s", sigBase64)
        }
        slack.Signature = sigBytes
        return mycrypto.ValidatePersonalSignature(slack.SignaturePayload, sigBytes, slack.Pubkey)
    }
    return errors.New("Signature not found in the slack message.")
}
