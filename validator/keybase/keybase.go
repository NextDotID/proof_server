package keybase

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/nextdotid/proof-server/types"
	mycrypto "github.com/nextdotid/proof-server/util/crypto"
	"github.com/nextdotid/proof-server/validator"
	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"
)

type Keybase validator.Base

const (
	VALIDATE_TEMPLATE = "^Prove myself: I'm 0x([0-9a-f]{66}) on NextID. Signature: (.*)"
	POST_TEMPLATE     = "Prove myself: I'm 0x%s on NextID. Signature: %%SIG_BASE64%%"
	URL               = "https://%s.keybase.pub/NextID/0x%s.txt"
)

var (
	l  = logrus.WithFields(logrus.Fields{"module": "validator", "validator": "keybase"})
	re = regexp.MustCompile(VALIDATE_TEMPLATE)
)

func Init() {
	if validator.PlatformFactories == nil {
		validator.PlatformFactories = make(map[types.Platform]func(validator.Base) validator.IValidator)
	}
	validator.PlatformFactories[types.Platforms.Keybase] = func(base validator.Base) validator.IValidator {
		kb := Keybase(base)
		return &kb
	}
}

func (kb *Keybase) GeneratePostPayload() (post string) {
	return fmt.Sprintf(POST_TEMPLATE, mycrypto.CompressedPubkeyHex(kb.Pubkey))
}

func (kb *Keybase) GenerateSignPayload() (payload string) {
	kb.Identity = strings.ToLower(kb.Identity)
	payloadStruct := validator.H{
		"action":   string(kb.Action),
		"identity": kb.Identity,
		"platform": "keybase",
		"prev":     nil,
	}
	if kb.Previous != "" {
		payloadStruct["prev"] = kb.Previous
	}

	payloadBytes, err := json.Marshal(payloadStruct)
	if err != nil {
		l.Warnf("Error when marshaling struct: %s", err.Error())
		return ""
	}

	return string(payloadBytes)
}

func (kb *Keybase) Validate() (err error) {
	kb.Identity = strings.ToLower(kb.Identity)
	url := fmt.Sprintf(URL, kb.Identity, mycrypto.CompressedPubkeyHex(kb.Pubkey))
	kb.ProofLocation = url
	resp, err := http.Get(url)
	if err != nil {
		return xerrors.Errorf("Error when requesting proof: %s", err.Error())
	}
	if resp.StatusCode != 200 {
		return xerrors.Errorf("Error when requesting proof: Status code %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return xerrors.Errorf("Error when getting resp body")
	}
	kb.Text = strings.TrimSpace(string(body))
	return kb.validateBody()
}

func (kb *Keybase) validateBody() error {
	l := l.WithFields(logrus.Fields{"function": "validateBody", "keybase": kb.Identity})
	matched := re.FindStringSubmatch(kb.Text)
	l.Debugf("Body: \"%s\"", kb.Text)
	if len(matched) < 3 {
		return xerrors.Errorf("Proof text struct mismatch.")
	}

	pubkeyHex := matched[1]
	pubkeyRecovered, err := mycrypto.StringToPubkey(pubkeyHex)
	if err != nil {
		return xerrors.Errorf("Pubkey recover failed: %s", err.Error())
	}
	if crypto.PubkeyToAddress(*kb.Pubkey) != crypto.PubkeyToAddress(*pubkeyRecovered) {
		return xerrors.Errorf("Pubkey mismatch")
	}

	sigBase64 := matched[2]
	sigBytes, err := base64.StdEncoding.DecodeString(sigBase64)
	if err != nil {
		return xerrors.Errorf("Error when decoding signature: %s", err.Error())
	}
	kb.Signature = sigBytes
	return mycrypto.ValidatePersonalSignature(kb.GenerateSignPayload(), sigBytes, pubkeyRecovered)
}
