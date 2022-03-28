package generate

import (
	"crypto/ecdsa"
	"github.com/nextdotid/proof-server/validator/ethereum"
	"github.com/nextdotid/proof-server/validator/github"
	"github.com/nextdotid/proof-server/validator/keybase"
	"github.com/nextdotid/proof-server/validator/twitter"
)

func init_validators() {
	twitter.Init()
	github.Init()
	keybase.Init()
	ethereum.Init()
}

func GeneratePayload() {
}

func generate_ethereum(sign_payload string, persona_sk *ecdsa.PrivateKey) {

}
