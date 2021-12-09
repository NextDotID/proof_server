package twitter

import (
	"encoding/base64"
	"encoding/json"
	"regexp"
	"strconv"
	"time"

	t "github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/ethereum/go-ethereum/common"
	"github.com/nextdotid/proof-server/config"
	"github.com/nextdotid/proof-server/util"
	"github.com/sirupsen/logrus"

	"github.com/nextdotid/proof-server/types"
)

type Twitter struct {
	// Previous signature hex ("0x.....")
	Previous string
	Action   types.Action
	Pubkey   common.Hash
	// Twitter screen name
	Identity string
	// TweetID
	ProofLocation string
	// Filled when tweet fetched successfully.
	TweetText string
}

const (
	TEMPLATE = "^Prove myself: I'm (0x[0-9a-f]{130}) on NextID. Signature: (.*)$"
)

var (
	client *t.Client
	l      = logrus.WithFields(logrus.Fields{"module": "validator", "validator": "twitter"})
	re     = regexp.MustCompile(TEMPLATE)
)

func (twitter *Twitter) GenerateSignPayload() (payload string) {
	now := time.Now().Unix()

	var payloadStruct map[string]interface{}
	payloadStruct = map[string]interface{}{
		"action":     string(twitter.Action),
		"platform":   "twitter",
		"identity":   twitter.Identity,
		"created_at": now,
		"prev":       nil,
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

func (twitter *Twitter) Validate() (result bool) {
	initClient()
	tweetID, err := strconv.ParseInt(twitter.ProofLocation, 10, 64)
	if err != nil {
		l.Warnf("Error when parsing tweet ID %s: %s", twitter.ProofLocation, err.Error())
		return false
	}

	tweet, _, err := client.Statuses.Show(tweetID, nil)
	if err != nil {
		l.Warnf("Error when getting tweet %s: %s", twitter.ProofLocation, err.Error())
		return false
	}
	if tweet.User.ScreenName != twitter.Identity {
		l.Warnf("Screen name mismatch: expect %s - actual %s", twitter.Identity, tweet.User.ScreenName)
		return false
	}

	twitter.TweetText = tweet.Text
	l.Debugf("Tweet text for %s: %s", twitter.ProofLocation, twitter.TweetText)

	return twitter.validateText()
}

func (twitter *Twitter) validateText() bool {
	l := l.WithFields(logrus.Fields{"function": "validateText", "tweet": twitter.ProofLocation})
	matched := re.FindStringSubmatch(twitter.TweetText)
	if len(matched) < 3 {
		l.Warnf("Tweet struct mismatch. Found: %+v", matched)
		return false
	}

	pubkeyHex := matched[1]
	if twitter.Pubkey.Hex() != pubkeyHex {
		return false
	}

	sigBase64 := matched[2]
	sigBytes, err := base64.StdEncoding.DecodeString(sigBase64)
	if err != nil {
		l.Warnf("Error when decoding signature %s: %s", sigBase64, err.Error())
		return false
	}
	sigHex := common.Bytes2Hex(sigBytes)
	return util.ValidatePersonalSignature(twitter.GenerateSignPayload(), sigHex, pubkeyHex)
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
