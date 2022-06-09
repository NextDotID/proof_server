package das

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/nextdotid/proof-server/types"
	"github.com/nextdotid/proof-server/util"
	mycrypto "github.com/nextdotid/proof-server/util/crypto"
	"github.com/nextdotid/proof-server/validator"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"
)

type Das struct {
	*validator.Base
}

type DasSignPayload struct {
	Version     string `json:"version"`
	Comment     string `json:"comment"`
	Comment2    string `json:"comment2"`
	Persona     string `json:"persona"`
	BitAddress  string `json:"bit_address"`
	SignPayload string `json:"sign_payload"`
	Signature   string `json:"signature"`
	CreatedAt   string `json:"created_at"`
	Uuid        string `json:"uuid"`
}

type DasRequest struct {
	Account string `json:"account"`
}

type DasResponse struct {
	ErrorNumber  int    `json:"err_no"`
	ErrorMessage string `json:"err_msg"`
	Data         struct {
		Records []DasRecord `json:"records"`
	}
}

type DasRecord struct {
	Key   string `json:"key"`
	Type  string `json:"type"`
	Label string `json:"label"`
	Value string `json:"value"`
	Ttl   string `json:"ttl"`
}

const (
	// v1 API.
	URL = "https://register-api.did.id/v1/account/records"
	Key = "nextid"
)

var (
	l = logrus.WithFields(logrus.Fields{"module": "validator", "validator": "dotbit"})
)

func Init() {
	if validator.PlatformFactories == nil {
		validator.PlatformFactories = make(map[types.Platform]func(*validator.Base) validator.IValidator)
	}
	validator.PlatformFactories[types.Platforms.Das] = func(base *validator.Base) validator.IValidator {
		das := Das{base}
		return &das
	}
}

func (das *Das) GeneratePostPayload() (_ map[string]string) {
	return map[string]string{"default": "%COMPRESSED_PERSONA_PUBKEY_HEX%:%SIG_BASE64%"}
}

func (das *Das) GenerateSignPayload() (payload string) {
	das.Identity = strings.ToLower(das.Identity)
	payloadStruct := validator.H{
		"action":     string(das.Action),
		"identity":   strings.ToLower(das.Identity),
		"platform":   string(types.Platforms.Das),
		"prev":       nil,
		"created_at": util.TimeToTimestampString(das.CreatedAt),
		"uuid":       das.Uuid.String(),
	}
	if das.Previous != "" {
		payloadStruct["prev"] = das.Previous
	}

	payloadBytes, err := json.Marshal(payloadStruct)
	if err != nil {
		l.Warnf("Error when marshaling struct: %s", err.Error())
		return ""
	}

	return string(payloadBytes)
}

func (das *Das) Validate() (err error) {
	das.Identity = strings.ToLower(das.Identity)
	das.SignaturePayload = das.GenerateSignPayload()

	// Find the record through API response instead of saving its 'location'.
	req, err := json.Marshal(DasRequest{Account: das.Identity})
	if err != nil {
		return xerrors.Errorf("Error when marshalling request: %w", err)
	}

	resp, err := http.Post(URL, "application/json", bytes.NewReader(req))
	if err != nil {
		return xerrors.Errorf("Error when requesting proof: %s", err.Error())
	}
	if resp.StatusCode != 200 {
		return xerrors.Errorf("Error when requesting proof: Status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return xerrors.New("Error when getting resp body")
	}

	result := new(DasResponse)
	err = json.Unmarshal(body, result)
	if err != nil {
		return xerrors.Errorf("error when decoding JSON: %w", err)
	}

	return das.validateRecord(result)
}

func (das *Das) validateRecord(resp *DasResponse) error {
	if resp.ErrorNumber != 0 {
		return xerrors.Errorf("err_no %d: %s", resp.ErrorNumber, resp.ErrorMessage)
	}
	// Colon as the separator between public key and signature.
	valuePrefix := "0x" + mycrypto.CompressedPubkeyHex(das.Pubkey) + ":"
	record, ok := lo.Find(resp.Data.Records, func(i DasRecord) bool {
		// Find the first record starts with the public key.
		return i.Key == Key && i.Type == "profile" && strings.HasPrefix(i.Value, valuePrefix)
	})
	if !ok {
		return xerrors.New("no key found")
	}

	_, sig, found := strings.Cut(record.Value, ":")
	if !found || len(sig) == 0 {
		return xerrors.New("invalid record value")
	}

	sigBytes, err := util.DecodeString(sig)
	if err != nil {
		return xerrors.Errorf("error when decoding sig: %w", err)
	}

	das.Signature = sigBytes
	return mycrypto.ValidatePersonalSignature(das.SignaturePayload, sigBytes, das.Pubkey)
}
