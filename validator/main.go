package validator

import (
	"crypto/ecdsa"

	"github.com/nextdotid/proof-server/types"
)

var (
	// Platforms contains all supported platform factory.
	Platforms map[types.Platform]func(Base) IValidator
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
	Platform      types.Platform
	Previous      string
	Action        types.Action
	Pubkey        *ecdsa.PublicKey
	// Identity on target platform.
	Identity      string
	ProofLocation string
	Signature     []byte
	Text          string
	// Extra info needed by separate platforms (e.g. Ethereum)
	Extra         map[string]string
}

// H for JSON builder.
type H map[string]interface{}
