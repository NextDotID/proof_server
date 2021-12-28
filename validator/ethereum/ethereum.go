package ethereum

import (
	"encoding/base64"
	"encoding/json"
	"regexp"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/nextdotid/proof-server/types"
	mycrypto "github.com/nextdotid/proof-server/util/crypto"
	"github.com/nextdotid/proof-server/validator"
	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"
)

type Ethereum validator.Base

const (
	VALIDATE_TEMPLATE = `^{"eth_address":"0x([0-9a-fA-F]{40})","signature":"(.*)"}$`
)

var (
	l  = logrus.WithFields(logrus.Fields{"module": "validator", "validator": "ethereum"})
	re = regexp.MustCompile(VALIDATE_TEMPLATE)
)

func Init() {
	if validator.PlatformFactories == nil {
		validator.PlatformFactories = make(map[types.Platform]func(validator.Base) validator.IValidator)
	}
	validator.PlatformFactories[types.Platforms.Ethereum] = func(base validator.Base) validator.IValidator {
		eth := Ethereum(base)
		return &eth
	}
}

// Not used by etheruem (for now).
func (*Ethereum) GeneratePostPayload() (post string) {
	return ""
}

func (et *Ethereum) GenerateSignPayload() (payload string) {
	payloadStruct := validator.H{
		"action":   string(et.Action),
		"identity": et.Identity,
		"persona":  "0x" + mycrypto.CompressedPubkeyHex(et.Pubkey),
		"platform": "ethereum",
		"prev":     nil,
	}
	if et.Previous != "" {
		payloadStruct["prev"] = et.Previous
	}
	payloadBytes, err := json.Marshal(payloadStruct)
	if err != nil {
		l.Warnf("Error when marshaling struct: %s", err.Error())
		return ""
	}
	return string(payloadBytes)
}

func (et *Ethereum) Validate() (err error) {
	// ETH wallet signature
	walletSignature, ok := et.Extra["wallet_signature"]
	if !ok {
		return xerrors.Errorf("wallet_signature not found")
	}
	if err := validateEthSignature(walletSignature, et.GenerateSignPayload(), et.Identity); err != nil {
		return xerrors.Errorf("%w", err)
	}

	// Persona signature
	return mycrypto.ValidatePersonalSignature(et.GenerateSignPayload(), et.Signature, et.Pubkey)
}

// `address` should be hexstring, `sig` should be BASE64-ed string.
func validateEthSignature(sig, payload, address string) error {
	addressGiven := common.HexToAddress(address)

	sigBytes, err := base64.StdEncoding.DecodeString(sig)
	if err != nil {
		return xerrors.Errorf("Error when decoding signature: %w", err)
	}

	pubkeyRecovered, err := mycrypto.RecoverPubkeyFromPersonalSignature(payload, sigBytes)
	if err != nil {
		return xerrors.Errorf("Error when extracting pubkey: %w", err.Error())
	}

	addressRecovered := crypto.PubkeyToAddress(*pubkeyRecovered)
	if addressRecovered.Hex() != addressGiven.Hex() {
		return xerrors.Errorf("ETH wallet signature validation failed")
	}
	return nil
}
