package twitter

import (
	"bufio"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	t "github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/nextdotid/proof-server/config"
	"github.com/nextdotid/proof-server/types"
	"github.com/nextdotid/proof-server/util"
	mycrypto "github.com/nextdotid/proof-server/util/crypto"
	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"

	"github.com/nextdotid/proof-server/validator"
)

type Twitter struct {
	*validator.Base
}

const (
	MATCH_TEMPLATE = "^Sig: (.*)$"
)

var (
	client      *t.Client
	l           = logrus.WithFields(logrus.Fields{"module": "validator", "validator": "twitter"})
	re          = regexp.MustCompile(MATCH_TEMPLATE)
	POST_STRUCT = map[string]string{
		"default": "🎭 Verifying my Twitter ID @%s for @NextDotID.\nSig: %%SIG_BASE64%%\n\nInstall Mask.io to enhance your Web3 experience.\n",
		"en_US":   "🎭 Verifying my Twitter ID @%s for @NextDotID.\nSig: %%SIG_BASE64%%\n\nInstall Mask.io to enhance your Web3 experience.\n",
		"zh_CN":   "🎭 正在通过 @NextDotID 验证我的 Twitter 帐号 @%s 。\nSig: %%SIG_BASE64%%\n\n请下载安装 Mask.io 去增强您的 Web3 体验。\n",
	}
)

func Init() {
	initClient()
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
		post[lang_code] = fmt.Sprintf(template, twitter.Identity)
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
	initClient()
	twitter.Identity = strings.ToLower(twitter.Identity)
	twitter.SignaturePayload = twitter.GenerateSignPayload()
	// Deletion. No need to fetch tweet.
	if twitter.Action == types.Actions.Delete {
		return mycrypto.ValidatePersonalSignature(twitter.SignaturePayload, twitter.Signature, twitter.Pubkey)
	}

	tweetID, err := strconv.ParseInt(twitter.ProofLocation, 10, 64)
	if err != nil {
		return xerrors.Errorf("Error when parsing tweet ID %s: %s", twitter.ProofLocation, err.Error())
	}

	tweet, _, err := client.Statuses.Show(tweetID, &t.StatusShowParams{
		TweetMode: "extended",
	})
	if err != nil {
		return xerrors.Errorf("Error when getting tweet %s: %w", twitter.ProofLocation, err)
	}
	if strings.ToLower(tweet.User.ScreenName) != strings.ToLower(twitter.Identity) {
		return xerrors.Errorf("Screen name mismatch: expect %s - actual %s", twitter.Identity, tweet.User.ScreenName)
	}

	twitter.Text = tweet.FullText
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
			return xerrors.Errorf("Error when decoding signature %s: %s", sigBase64, err.Error())
		}
		twitter.Signature = sigBytes
		return mycrypto.ValidatePersonalSignature(twitter.SignaturePayload, sigBytes, twitter.Pubkey)
	}
	return xerrors.Errorf("Signature not found in tweet text.")
}

func initClient() {
	if client != nil {
		return
	}
	oauthToken := oauth1.NewToken(
		config.C.Platform.Twitter.AccessToken,
		config.C.Platform.Twitter.AccessTokenSecret,
	)
	oauthConfig := oauth1.NewConfig(
		config.C.Platform.Twitter.ConsumerKey,
		config.C.Platform.Twitter.ConsumerSecret,
	)
	httpClient := oauthConfig.Client(oauth1.NoContext, oauthToken)
	client = t.NewClient(httpClient)
}
