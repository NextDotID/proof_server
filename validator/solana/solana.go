package solana

import (
	"encoding/json"

	"github.com/gagliardetto/solana-go"
	"github.com/mr-tron/base58"
	"github.com/nextdotid/proof-server/types"
	"github.com/nextdotid/proof-server/util"
	mycrypto "github.com/nextdotid/proof-server/util/crypto"
	"github.com/nextdotid/proof-server/validator"
	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"
)

type Solana struct {
	*validator.Base
}

var (
	l = logrus.WithFields(logrus.Fields{"module": "validator", "validator": "solana"})
)

func Init() {
	if validator.PlatformFactories == nil {
		validator.PlatformFactories = make(map[types.Platform]func(*validator.Base) validator.IValidator)
	}
	validator.PlatformFactories[types.Platforms.Solana] = func(base *validator.Base) validator.IValidator {
		sol := Solana{base}
		return &sol
	}
}

func (*Solana) GeneratePostPayload() (post map[string]string) {
	return map[string]string{"default": ""}
}

func (sol *Solana) GenerateSignPayload() (payload string) {
	payloadStruct := validator.H{
		"action":     string(sol.Action),
		"identity":   sol.Identity,
		"persona":    "0x" + mycrypto.CompressedPubkeyHex(sol.Pubkey),
		"platform":   "solana",
		"prev":       nil,
		"created_at": util.TimeToTimestampString(sol.CreatedAt),
		"uuid":       sol.Uuid.String(),
	}
	if sol.Previous != "" {
		payloadStruct["prev"] = sol.Previous
	}
	payloadBytes, err := json.Marshal(payloadStruct)
	if err != nil {
		l.Warnf("Error when marshalling struct: %s", err.Error())
		return ""
	}
	return string(payloadBytes)
}

func (sol *Solana) Validate() (err error) {
	// Wallet Sig encoded by Base58
	// Persona Sig encoded by Base64
	sol.SignaturePayload = sol.GenerateSignPayload()

	switch sol.Action {
	case types.Actions.Create:
		{
			return sol.validateCreate()
		}
	case types.Actions.Delete:
		{
			return sol.validateDelete()
		}
	default:
		{
			return xerrors.Errorf("unknown action: %s", sol.Action)
		}
	}
}

func (sol *Solana) validateCreate() (err error) {
	walletSig, ok := sol.Extra["wallet_signature"]
	if !ok {
		return xerrors.Errorf("wallet_signature not found")
	}

	if err := validateWalletSignature(sol.SignaturePayload, walletSig, sol.Identity); err != nil {
		return xerrors.Errorf("invalid wallet signature %w", err)
	}

	if err := mycrypto.ValidatePersonalSignature(sol.SignaturePayload, sol.Signature, sol.Pubkey); err != nil {
		return xerrors.Errorf("invalid persona signature %w", err)
	}

	return nil
}

func (sol *Solana) validateDelete() (err error) {
	walletSig, ok := sol.Extra["wallet_signature"]

	// If wallet_signature exists, check it
	if ok && walletSig != "" {
		err := validateWalletSignature(sol.SignaturePayload, walletSig, sol.Identity)
		if err != nil {
			return xerrors.Errorf("invalid wallet signature %w", err)
		}

		sigBytes, err := base58.Decode(walletSig)
		if err != nil {
			return xerrors.Errorf("invalid wallet signature format %w", err)
		}

		sol.Signature = sigBytes
		return nil
	}

	if err := mycrypto.ValidatePersonalSignature(sol.SignaturePayload, sol.Signature, sol.Pubkey); err != nil {
		return xerrors.Errorf("invalid persona signature %w", err)
	}

	return nil
}

func validateWalletSignature(payload, sig, address string) error {
	pubkey, err := solana.PublicKeyFromBase58(address)
	if err != nil {
		return xerrors.Errorf("error when decoding pubkey: %w", err)
	}

	signature, err := solana.SignatureFromBase58(sig)
	if err != nil {
		return xerrors.Errorf("error when decoding signature: %w", err)
	}
	if !signature.Verify(pubkey, []byte(payload)) {
		return xerrors.Errorf("solana wallet signature validation failed")
	}
	return nil
}
