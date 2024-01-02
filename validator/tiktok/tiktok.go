package tiktok

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/nextdotid/proof_server/types"
	"github.com/nextdotid/proof_server/util"
	mycrypto "github.com/nextdotid/proof_server/util/crypto"
	"github.com/nextdotid/proof_server/validator"
	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"
)

type TikTok struct {
	*validator.Base
}

const (
	MATCH_TEMPLATE = `Sig: (.+?)[\s\n;]`
	MATCH_MISC     = `\bMisc: ([^|]+)\|(\d+)\|(.+)?$`
)

var (
	l           = logrus.WithFields(logrus.Fields{"module": "validator", "validator": "tiktok"})
	POST_STRUCT = map[string]string{
		"default": `🎭 NextID ROCKS! Sig: %%SIG_BASE64%%;Misc: %s|%s|%s`,
	}
	re = regexp.MustCompile(MATCH_TEMPLATE)
)

func Init() {
	if validator.PlatformFactories == nil {
		validator.PlatformFactories = make(map[types.Platform]func(*validator.Base) validator.IValidator)
	}
	validator.PlatformFactories[types.Platforms.TikTok] = func(base *validator.Base) validator.IValidator {
		tt := TikTok{base}
		return &tt
	}
}

func (tt *TikTok) GeneratePostPayload() (post map[string]string) {
	post = make(map[string]string, 0)
	for lang_code, template := range POST_STRUCT {
		post[lang_code] = fmt.Sprintf(template, tt.Identity, tt.Uuid.String(), util.TimeToTimestampString(tt.CreatedAt), tt.Previous)
	}

	return post
}

func (tt *TikTok) GenerateSignPayload() (payload string) {
	tt.Identity = strings.ToLower(tt.Identity)
	payloadStruct := validator.H{
		"action":     string(tt.Action),
		"identity":   tt.Identity,
		"platform":   types.Platforms.TikTok,
		"prev":       nil,
		"created_at": util.TimeToTimestampString(tt.CreatedAt),
		"uuid":       tt.Uuid.String(),
	}
	if tt.Previous != "" {
		payloadStruct["prev"] = tt.Previous
	}

	payloadBytes, err := json.Marshal(payloadStruct)
	if err != nil {
		l.Warnf("Error when marshaling struct: %s", err.Error())
		return ""
	}

	return string(payloadBytes)
}

func (tt *TikTok) Validate() (err error) {
	oembedInfo, err := fetchOembedInfo(tt.ProofLocation)
	if err != nil {
		return xerrors.Errorf("error when fetching tiktok proof: %w", err)
	}
	if  tt.Identity != oembedInfo.AuthorUniqueID {
		return xerrors.Errorf("tiktok user mismatch: %s instead of %s", oembedInfo.AuthorUniqueID, tt.Identity)
	}
	signature, err := extractSignatureFromTitle(oembedInfo.Title)
	if err != nil {
		return err
	}
	tt.Signature = signature
	tt.ProofLocation = oembedInfo.EmbedProductID
	return mycrypto.ValidatePersonalSignature(tt.GenerateSignPayload(), tt.Signature, tt.Pubkey)
}

func (tt *TikTok) GetAltID() string {
	return ""
}

func extractSignatureFromTitle(title string) (signature []byte, err error) {
	result := re.FindStringSubmatch(title)
	if len(result) != 2 {
		return []byte{}, xerrors.New("signature not found in tiktok title")
	}
	signature, err = base64.StdEncoding.DecodeString(result[1])
	if err != nil {
		return []byte{}, xerrors.Errorf("when decoding tiktok signature: %w", err)
	}

	return signature, nil
}
