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
	Previous      string
	Action        types.Action
	Pubkey        *ecdsa.PublicKey
	Identity      string
	ProofLocation string
}
