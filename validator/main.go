package validator

import (
	"crypto/ecdsa"

	"github.com/nextdotid/proof-server/types"
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
	Identity      string
	ProofLocation string
	Signature     []byte
	Text string
}

// H for JSON builder.
type H map[string]interface{}
