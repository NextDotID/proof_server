package twitter

import (
	"bufio"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/nextdotid/proof_server/types"
	"github.com/nextdotid/proof_server/util"
	mycrypto "github.com/nextdotid/proof_server/util/crypto"
	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"

	"github.com/nextdotid/proof_server/validator"
)

type Twitter struct {
	*validator.Base
}

const (
	MATCH_TEMPLATE = "\\bSig: (.+?)[\\s\\n]"
)

var (
	l           = logrus.WithFields(logrus.Fields{"module": "validator", "validator": "twitter"})
	re          = regexp.MustCompile(MATCH_TEMPLATE)
	POST_STRUCT = map[string]string{
		// Misc info: UUID|CreatedAt|Previous
		"default": "🎭 Verifying my Twitter ID @%s for @NextDotID.\nSig: %%SIG_BASE64%%\nMisc info: %s|%s\n\nPowered by Next.ID - Connect All Digital Identities.\n",
		"en_US":   "🎭 Verifying my Twitter ID @%s for @NextDotID.\nSig: %%SIG_BASE64%%\nMisc info: %s|%s\n\nPowered by Next.ID - Connect All Digital Identities.\n",
		"zh_CN":   "🎭 正在通过 @NextDotID 验证我的 Twitter 帐号 @%s 。\nSig: %%SIG_BASE64%%\n其它信息: %s|%s\n\n由 Next.ID 支持 - 连接全域数字身份。\n",
	}
)

func Init() {
	if validator.PlatformFactories == nil {
		validator.PlatformFactories = make(map[types.Platform]func(*validator.Base) validator.IValidator)
	}

	validator.PlatformFactories[types.Platforms.Twitter] = func(base *validator.Base) validator.IValidator {
		twi := Twitter{base}
		return &twi
	}
}

func (twitter *Twitter) GeneratePostPayload() (post map[string]string) {
	post = make(map[string]string, 0)
	for lang_code, template := range POST_STRUCT {
		post[lang_code] = fmt.Sprintf(template, twitter.Identity, twitter.Uuid.String(), util.TimeToTimestampString(twitter.CreatedAt))
	}

	return post
}

func (twitter *Twitter) GenerateSignPayload() (payload string) {
	twitter.Identity = strings.ToLower(twitter.Identity)
	payloadStruct := validator.H{
		"action":     string(twitter.Action),
		"identity":   twitter.Identity,
		"platform":   "twitter",
		"prev":       nil,
		"created_at": util.TimeToTimestampString(twitter.CreatedAt),
		"uuid":       twitter.Uuid.String(),
	}
	if twitter.Previous != "" {
		payloadStruct["prev"] = twitter.Previous
	}

	payloadBytes, err := json.Marshal(payloadStruct)
	if err != nil {
		l.Warnf("Error when marshaling struct: %s", err.Error())
		return ""
	}

	return string(payloadBytes)
}

func (twitter *Twitter) Validate() (err error) {
	twitter.Identity = strings.ToLower(twitter.Identity)
	if twitter.SignaturePayload == "" {
		twitter.SignaturePayload = twitter.GenerateSignPayload()
	}

	// Deletion. No need to fetch tweet.
	if twitter.Action == types.Actions.Delete {
		return mycrypto.ValidatePersonalSignature(twitter.SignaturePayload, twitter.Signature, twitter.Pubkey)
	}

	tweetID, err := strconv.ParseInt(twitter.ProofLocation, 10, 64)
	if err != nil {
		return xerrors.Errorf("parsing tweet ID %s: %s", twitter.ProofLocation, err.Error())
	}

	post, err := validator.GetPostWithHeadlessBrowser(
		fmt.Sprintf("https://twitter.com/%s/status/%d", twitter.Identity, tweetID),
		"Sig:",
	)
	if err != nil {
		return xerrors.Errorf("fetching tweet with headless browser: %w", err)
	}

	twitter.Text = post
	return twitter.validateText()
}

func (twitter *Twitter) validateText() (err error) {
	scanner := bufio.NewScanner(strings.NewReader(twitter.Text))
	for scanner.Scan() {
		matched := re.FindStringSubmatch(scanner.Text())
		if len(matched) < 2 {
			continue // Search for next line
		}

		sigBase64 := matched[1]
		sigBytes, err := util.DecodeString(sigBase64)
		if err != nil {
			return xerrors.Errorf("decoding signature %s: %s", sigBase64, err.Error())
		}
		twitter.Signature = sigBytes
		return mycrypto.ValidatePersonalSignature(twitter.SignaturePayload, sigBytes, twitter.Pubkey)
	}
	return xerrors.Errorf("Signature not found in tweet text.")
}
