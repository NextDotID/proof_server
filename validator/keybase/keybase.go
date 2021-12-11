package keybase

import (
	"crypto/ecdsa"
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
	"github.com/sirupsen/logrus"
)

type Keybase struct {
	Previous string
	Action types.Action
	Pubkey *ecdsa.PublicKey
	Identity string
	ProofLocation string
	ProofText string
}

const (
	TEMPLATE = "^Prove myself: I'm 0x([0-9a-f]{66}) on NextID. Signature: (.*)"
	URL = "https://%s.keybase.pub/NextID/0x%s.txt"
)

var (
	l = logrus.WithFields(logrus.Fields{"module": "validator", "validator": "keybase"})
	re = regexp.MustCompile(TEMPLATE)
)

func (kb *Keybase) GenerateSignPayload() (payload string) {
	var payloadStruct map[string]interface{}
	payloadStruct = map[string]interface{}{
		"action":   string(kb.Action),
		"platform": "keybase",
		"identity": kb.Identity,
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

func (kb *Keybase) Validate() (result bool) {
	url := fmt.Sprintf(URL, kb.Identity, mycrypto.CompressedPubkeyHex(kb.Pubkey))
	kb.ProofLocation = url
	resp, err := http.Get(url)
	if err != nil {
		l.Warnf("Error when requesting proof: %s", err.Error())
		return false
	}
	if resp.StatusCode != 200 {
		l.Warnf("Error when requesting proof: Status code %d", resp.StatusCode)
		return false
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		l.Warnf("Error when getting resp body")
		return false
	}
	kb.ProofText = strings.TrimSpace(string(body))
	return kb.validateBody()
}

func (kb *Keybase) validateBody() bool {
	l := l.WithFields(logrus.Fields{"function": "validateBody", "keybase": kb.Identity})
	matched := re.FindStringSubmatch(kb.ProofText)
	l.Debugf("Body: \"%s\"", kb.ProofText)
	if len(matched) < 3 {
		l.Warnf("Proof text struct mismatch. Found: %+v", matched)
	}

	pubkeyHex := matched[1]
	pubkeyRecovered, err := mycrypto.StringToPubkey(pubkeyHex)
	if err != nil {
		l.Warnf("Pubkey recover failed: %s", err.Error())
		return false
	}
	if crypto.PubkeyToAddress(*kb.Pubkey) != crypto.PubkeyToAddress(*pubkeyRecovered) {
		l.Warnf("Pubkey mismatch")
		return false
	}

	sigBase64 := matched[2]
	sigBytes, err := base64.StdEncoding.DecodeString(sigBase64)
	if err != nil {
		l.Warnf("Error when decoding signature %s: %s", sigBase64, err.Error())
		return false
	}
	return mycrypto.ValidatePersonalSignature(kb.GenerateSignPayload(), sigBytes, pubkeyRecovered)
}
