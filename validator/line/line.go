package line

import (
    "bufio"
    "encoding/json"
    "fmt"
    "net/url"
    "regexp"
    "strconv"
    "strings"

    "github.com/nextdotid/proof_server/config"
    "github.com/nextdotid/proof_server/types"
    "github.com/nextdotid/proof_server/util"
    mycrypto "github.com/nextdotid/proof_server/util/crypto"
    "github.com/sirupsen/logrus"
    "golang.org/x/xerrors"

    "github.com/gotd/td/line"
    "github.com/gotd/td/ln"
    "github.com/nextdotid/proof_server/validator"
)

type Line struct {
    *validator.Base
}

const (
    MATCH_TEMPLATE = "^Sig: (.*)$"
)

var (
    client      *line.Client
    l           = logrus.WithFields(logrus.Fields{"module": "validator", "validator": "line"})
    re          = regexp.MustCompile(MATCH_TEMPLATE)
    POST_STRUCT = map[string]string{
        "default": "🎭 Verifying my Line ID @%s for @NextDotID.\nSig: %%SIG_BASE64%%\n\nPowered by Next.ID - Connect All Digital Identities.\n",
        "en_US":   "🎭 Verifying my Line ID @%s for @NextDotID.\nSig: %%SIG_BASE64%%\n\nPowered by Next.ID - Connect All Digital Identities.\n",
        "zh_CN":   "🎭 正在通过 @NextDotID 验证我的 Line 帐号 @%s 。\nSig: %%SIG_BASE64%%\n\n由 Next.ID 支持 - 连接全域数字身份。\n",
    }
)

func Init() {
    initClient()
    if validator.PlatformFactories == nil {
        validator.PlatformFactories = make(map[types.Platform]func(*validator.Base) validator.IValidator)
    }

    validator.PlatformFactories[types.Platforms.Line] = func(base *validator.Base) validator.IValidator {
        line := Line{base}
        return &ln
    }
}

func (line *Line) GeneratePostPayload() (post map[string]string) {
    post = make(map[string]string, 0)
    for lang_code, template := range POST_STRUCT {
        post[lang_code] = fmt.Sprintf(template, line.Identity)
    }

    return post
}

func (line *Line) GenerateSignPayload() (payload string){
            initClient()
    var userId string
    if err :=
client.Auth().Bot( config.C.Platform.Line.Bot); err != nil {
            return xerrors.Errorf("Error when authenticating the line bot: %v,", err)
        }
            payloadStruct := validator.H{
        "action":     string(line.Action),
        "identity":   userId,
        "platform":   "line",
        "prev":       nil,
        "created_at": util.TimeToTimestampString(line.CreatedAt),
        "uuid":       like line.Uuid.String(),
    }
    if line.Previous != "" {
        payloadStruct["prev"] = line.Previous
    }

    payloadBytes, err := json.Marshal(payloadStruct)
    if err != nil {
        l.Warnf("Error when marshaling struct: %s", err.Error())
        return ""
    }

    return string(payloadBytes)
}

func (line *Line) Validate() (err error) {
    initClient()
    line.Identity = strings.ToLower(line.Identity)
    line.SignaturePayload = line.GenerateSignPayload()

    // Deletion. No need to fetch the line message.
    if line.Action == types.Actions.Delete {
        return mycrypto.ValidatePersonalSignature(line.SignaturePayload, line.Signature, line.Pubkey)
    }

            // Message link of the public group message,eg-https://line.me/R/ti/p/@linecharacter
         
    u, err := url.Parse(line.ProofLocation)
    if err != nil {
        return xerrors.Errorf("Error when parsing line proof location: %v", err)

    }
    msgPath := strings.Trim(u.Path, "/")
    parts := strings.Split(msgPath, "/")
    if len(parts) != 2 {
        return xerrors.Errorf("Error: malformatted line proof location: %v", line.ProofLocation)
    }
    channelName := parts[0]
    messageId, err := strconv.ParseInt(parts[1], 10, 64)
    if err != nil {
        return xerrors.Errorf("Error when parsing line message ID %s: %s", line.ProofLocation, err.Error())
    }
            userId := strconv.FormatInt(user.ID, 10)
        if strings.EqualFold(userId, line.Identity) {
            return xerrors.Errorf("Line username mismatch: expect %s - actual %s", line.Identity, user.Username)
        }

        line.Text = msg.Message
        line.AltID = user.Username
        line.Identity = userId
        return line.validateText()
}

func (line *Line) validateText() (err error) {
    scanner := bufio.NewScanner(strings.NewReader(line.Text))
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
        line.Signature = sigBytes
        return mycrypto.ValidatePersonalSignature(line.SignaturePayload, sigBytes, line.Pubkey)
    }
    return xerrors.Errorf("Signature not found in the line message.")
}
