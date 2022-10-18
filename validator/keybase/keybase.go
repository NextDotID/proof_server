package keybase

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/nextdotid/proof_server/types"
	"github.com/nextdotid/proof_server/util"
	mycrypto "github.com/nextdotid/proof_server/util/crypto"
	"github.com/nextdotid/proof_server/validator"
	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"
)

type Keybase struct {
	*validator.Base
}

type KeybasePayload struct {
	Version         string `json:"version"`
	Comment         string `json:"comment"`
	Comment2        string `json:"comment2"`
	Persona         string `json:"persona"`
	KeybaseUsername string `json:"keybase_username"`
	SignPayload     string `json:"sign_payload"`
	Signature       string `json:"signature"`
	CreatedAt       string `json:"created_at"`
	Uuid            string `json:"uuid"`
}

const (
	URL = "https://%s.keybase.pub/NextID/0x%s.json"
)

var (
	l = logrus.WithFields(logrus.Fields{"module": "validator", "validator": "keybase"})
)

func Init() {
	if validator.PlatformFactories == nil {
		validator.PlatformFactories = make(map[types.Platform]func(*validator.Base) validator.IValidator)
	}
	validator.PlatformFactories[types.Platforms.Keybase] = func(base *validator.Base) validator.IValidator {
		kb := Keybase{base}
		return &kb
	}
}

func (kb *Keybase) GeneratePostPayload() (post map[string]string) {
	kb.Identity = strings.ToLower(kb.Identity)
	payload := KeybasePayload{
		Version:         "1",
		Comment:         "Here's an NextID proof of this Keybase account.",
		Comment2:        "To validate, base64.decode the signature, and recover pubkey from it using sign_payload with ethereum personal_sign algo.",
		Persona:         "0x" + mycrypto.CompressedPubkeyHex(kb.Pubkey),
		KeybaseUsername: kb.Identity,
		SignPayload:     kb.GenerateSignPayload(),
		Signature:       "%SIG_BASE64%",
		CreatedAt:       util.TimeToTimestampString(kb.CreatedAt),
		Uuid:            kb.Uuid.String(),
	}
	payload_json, _ := json.MarshalIndent(payload, "", "\t")
	return map[string]string{"default": string(payload_json)}
}

func (kb *Keybase) GenerateSignPayload() (payload string) {
	kb.Identity = strings.ToLower(kb.Identity)
	payloadStruct := validator.H{
		"action":     string(kb.Action),
		"identity":   kb.Identity,
		"platform":   string(types.Platforms.Keybase),
		"prev":       nil,
		"created_at": util.TimeToTimestampString(kb.CreatedAt),
		"uuid":       kb.Uuid.String(),
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
	kb.SignaturePayload = kb.GenerateSignPayload()
	kb.AltID = kb.Identity // TODO: maybe get Keybase UserID in another API call?

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

	payload := new(KeybasePayload)
	err = json.Unmarshal(body, payload)
	if err != nil {
		return xerrors.Errorf("error when decoding JSON: %w", err)
	}
	return kb.validateBody(payload)
}

func (kb *Keybase) validateBody(payload *KeybasePayload) error {
	if payload.Persona != ("0x" + mycrypto.CompressedPubkeyHex(kb.Pubkey)) {
		return xerrors.Errorf("Persona mismatch")
	}

	sig_bytes, err := util.DecodeString(payload.Signature)
	if err != nil {
		return xerrors.Errorf("error when decoding sig: %w", err)
	}

	kb.Signature = sig_bytes
	return mycrypto.ValidatePersonalSignature(kb.SignaturePayload, sig_bytes, kb.Pubkey)
}
