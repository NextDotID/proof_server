package instagram

import (
	"bufio"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	t "github.com/dghubble/go-instagram/instagram"
	"github.com/dghubble/oauth1"
	"github.com/nextdotid/proof_server/config"
	"github.com/nextdotid/proof_server/types"
	"github.com/nextdotid/proof_server/util"
	mycrypto "github.com/nextdotid/proof_server/util/crypto"
	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"

	"github.com/nextdotid/proof_server/validator"
)

type instagram struct {
	*validator.Base
}

const (
	MATCH_TEMPLATE = "^Sig: (.*)$"
)

var (
	client      *t.Client
	l           = logrus.WithFields(logrus.Fields{"module": "validator", "validator": "instagram"})
	re          = regexp.MustCompile(MATCH_TEMPLATE)
	POST_STRUCT = map[string]string{
		"default": "🎭 Verifying my Instagram ID @%s for @NextDotID.\nSig: %%SIG_BASE64%%\n\nPowered by Next.ID - Connect All Digital Identities.\n",
		"en_US":   "🎭 Verifying my Instagram ID @%s for @NextDotID.\nSig: %%SIG_BASE64%%\n\nPowered by Next.ID - Connect All Digital Identities.\n",
		"zh_CN":   "🎭 正在通过 @NextDotID 验证我的 Instagram 帐号 @%s 。\nSig: %%SIG_BASE64%%\n\n由 Next.ID 支持 - 连接全域数字身份。\n",
	}
)

func Init() {
	initClient()
	if validator.PlatformFactories == nil {
		validator.PlatformFactories = make(map[types.Platform]func(*validator.Base) validator.IValidator)
	}

	validator.PlatformFactories[types.Platforms.Instagram] = func(base *validator.Base) validator.IValidator {
		instagram := Instagram{base}
		return &instagram
	}
}

func (instagram *Instagram) GeneratePostPayload() (post map[string]string) {
	post = make(map[string]string, 0)
	for lang_code, template := range POST_STRUCT {
		post[lang_code] = fmt.Sprintf(template, instagram.Identity)
	}

	return post
}

func (instagram *Instagram) GenerateSignPayload() (payload string) {
	instagram.Identity = strings.ToLower(instagram.Identity)
	payloadStruct := validator.H{
		"action":     string(instagram.Action),
		"identity":   instagram.Identity,
		"platform":   "instagram",
		"prev":       nil,
		"created_at": util.TimeToTimestampString(instagram.CreatedAt),
		"uuid":       instagram.Uuid.String(),
	}
	if instagram.Previous != "" {
		payloadStruct["prev"] = instagram.Previous
	}

	payloadBytes, err := json.Marshal(payloadStruct)
	if err != nil {
		l.Warnf("Error when marshaling struct: %s", err.Error())
		return ""
	}

	return string(payloadBytes)
}

func (instagram *Instagram) Validate() (err error) {
	initClient()
	instagram.Identity = strings.ToLower(instagram.Identity)
	instagram.SignaturePayload = instagram.GenerateSignPayload()
	// Deletion. No need to fetch post.
	if instagram.Action == types.Actions.Delete {
		return mycrypto.ValidatePersonalSignature(instagram.SignaturePayload, instagram.Signature, instagram.Pubkey)
	}

	postID, err := strconv.ParseInt(instagram.ProofLocation, 10, 64)
	if err != nil {
		return xerrors.Errorf("Error when parsing post ID %s: %s", instagram.ProofLocation, err.Error())
	}

	post, _, err := client.Statuses.Show(postID, &t.StatusShowParams{
		PostMode: "extended",
	})
	if err != nil {
		return xerrors.Errorf("Error when getting post %s: %w", instagram.ProofLocation, err)
	}
	if strings.ToLower(post.User.ScreenName) != strings.ToLower(instagram.Identity) {
		return xerrors.Errorf("Screen name mismatch: expect %s - actual %s", instagram.Identity, post.User.ScreenName)
	}

	instagram.Text = post.FullText
	instagram.AltID = strconv.FormatInt(post.User.ID, 10)
	return instagram.validateText()
}

func (instagram *Instagram) validateText() (err error) {
	scanner := bufio.NewScanner(strings.NewReader(post.Text))
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
		instagram.Signature = sigBytes
		return mycrypto.ValidatePersonalSignature(instagram.SignaturePayload, sigBytes, instagram.Pubkey)
	}
	return xerrors.Errorf("Signature not found in post text.")
}

func initClient() {
	if client != nil {
		return
	}
	oauthToken := oauth1.NewToken(
		config.C.Platform.Instagram.AccessToken,
		config.C.Platform.Instagram.PageID,
	)
	httpClient := oauthConfig.Client(oauth1.NoContext, oauthToken)
	client = t.NewClient(httpClient)
}
