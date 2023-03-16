package validator

import (
	"crypto/ecdsa"
	"time"

	"github.com/google/uuid"
	"github.com/nextdotid/proof_server/types"
)

var (
	// PlatformFactories contains all supported platform factory.
	PlatformFactories map[types.Platform]func(*Base) IValidator
)

type IValidator interface {
	// GeneratePostPayload gives a post structure (with
	// placeholders) for user to post on target platform.
	GeneratePostPayload() (post map[string]string)
	// GenerateSignPayload generates a string to be signed.
	GenerateSignPayload() (payload string)

	Validate() (err error)
}

type Base struct {
	Platform types.Platform
	Previous string
	Action   types.Action
	Pubkey   *ecdsa.PublicKey
	// Identity on target platform.
	Identity         string
	AltID            string
	ProofLocation    string
	Signature        []byte
	SignaturePayload string
	Text             string
	// Extra info needed by separate platforms (e.g. Ethereum)
	Extra map[string]string
	// CreatedAt indicates creation time of this link
	CreatedAt time.Time
	// Uuid gives this link an unique identifier, to let other
	// third-party service distinguish / store / dedup links with
	// ease.
	Uuid uuid.UUID
}

// BaseToInterface converts a `validator.Base` struct to
// `validator.IValidator` interface.
func BaseToInterface(v *Base) IValidator {
	performer_factory, ok := PlatformFactories[v.Platform]
	if !ok {
		return nil
	}

	return performer_factory(v)
}

// H for JSON builder.
type H map[string]any
