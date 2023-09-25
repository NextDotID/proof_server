package nextid

import (
	"encoding/json"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/nextdotid/proof_server/types"
	"github.com/nextdotid/proof_server/util"
	mycrypto "github.com/nextdotid/proof_server/util/crypto"
	"github.com/nextdotid/proof_server/validator"
	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"
)

type NextID struct {
	*validator.Base
}

var (
	l = logrus.WithFields(logrus.Fields{"module": "validator", "validator": "nextid"})
)

func Init() {
	if validator.PlatformFactories == nil {
		validator.PlatformFactories = make(map[types.Platform]func(*validator.Base) validator.IValidator)
	}

	validator.PlatformFactories[types.Platforms.NextID] = func(base *validator.Base) validator.IValidator {
		nextID := NextID{base}
		return &nextID
	}
}

func (nextID *NextID) GetAltID() (altID string) {
	return ""
}

func (nextID *NextID) GeneratePostPayload() (post map[string]string) {
	return map[string]string{
		"default": "",
	}
}

// GenerateSignPayload generates a string to be signed.  If empty, an error is occured internally.
func (nextID *NextID) GenerateSignPayload() (payload string) {
	targetAvatar, err := mycrypto.StringToPubkey(nextID.Identity)
	if err != nil {
		return ""
	}

	payloadStruct := validator.H{
		"action":     string(nextID.Action),
		"identity":   "0x" + mycrypto.CompressedPubkeyHex(targetAvatar),
		"persona":    "0x" + mycrypto.CompressedPubkeyHex(nextID.Pubkey),
		"platform":   "nextid",
		"prev":       nil,
		"created_at": util.TimeToTimestampString(nextID.CreatedAt),
		"uuid":       nextID.Uuid.String(),
	}
	if nextID.Previous != "" {
		payloadStruct["prev"] = nextID.Previous
	}
	payloadBytes, err := json.Marshal(payloadStruct)
	if err != nil {
		l.Warnf("Error when marshaling struct: %s", err.Error())
		return ""
	}
	return string(payloadBytes) // TODO
}

func (nextID *NextID) Validate() (err error) {
	targetSig, ok := nextID.Extra["target_signature"]
	if !ok {
		return xerrors.Errorf("Target Avatar signature not provided")
	}
	targetSigParsed := strings.TrimPrefix(targetSig, "0x")
	targetSigParsed = strings.ToLower(targetSigParsed)
	targetSigBytes := common.Hex2Bytes(targetSigParsed)

	targetAvatar, err := mycrypto.StringToPubkey(nextID.Identity)
	if err != nil {
		return xerrors.Errorf("Invalid target avatar: %s", nextID.Identity)
	}

	hexutil.Decode(targetSig)
	payload := nextID.GenerateSignPayload()
	if err := mycrypto.ValidatePersonalSignature(payload, nextID.Signature, nextID.Pubkey); err != nil {
		return xerrors.Errorf("Invalid base signature: %w", err)
	}
	if err := mycrypto.ValidatePersonalSignature(payload, targetSigBytes, targetAvatar); err != nil {
		return xerrors.Errorf("Invalid target signature: %w", err)
	}

	return nil // TODO
}
