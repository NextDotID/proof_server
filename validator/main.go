package validator

import (
	"crypto/ecdsa"

	"github.com/nextdotid/proof-server/types"
)

var (
	// PlatformFactories contains all supported platform factory.
	PlatformFactories map[types.Platform]func(*Base) IValidator
)

type IValidator interface {
	// GeneratePostPayload gives a post structure (with
	// placeholders) for user to post on target platform.
	GeneratePostPayload() (post string)
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
	ProofLocation    string
	Signature        []byte
	SignaturePayload string
	Text             string
	// Extra info needed by separate platforms (e.g. Ethereum)
	Extra map[string]string
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
type H map[string]interface{}
