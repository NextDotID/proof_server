package ethereum

import (
	"crypto/ecdsa"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/nextdotid/proof-server/config"
	"github.com/nextdotid/proof-server/types"

	mycrypto "github.com/nextdotid/proof-server/util/crypto"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var (
	persona_sk *ecdsa.PrivateKey
	wallet_sk  *ecdsa.PrivateKey

	eth = Ethereum{
		Platform: types.Platforms.Ethereum,
		Previous: "",
		Action:   types.Actions.Create,
		Extra: map[string]string{
			"wallet_signature": "",
		},
	}
)

func before_each(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	config.Init("../../config/config.test.json")

	_, persona_sk = mycrypto.GenerateKeypair()
	eth.Pubkey = &persona_sk.PublicKey

	_, wallet_sk = mycrypto.GenerateKeypair()
	eth.Identity = crypto.PubkeyToAddress(wallet_sk.PublicKey).Hex()

	// Generate sig
	eth.Signature, _ = mycrypto.SignPersonal([]byte(eth.GenerateSignPayload()), persona_sk)
	wallet_sig, _ := mycrypto.SignPersonal([]byte(eth.GenerateSignPayload()), wallet_sk)
	eth.Extra = map[string]string{
		"wallet_signature": base64.StdEncoding.EncodeToString(wallet_sig),
	}
}

func Test_GeneratePostPayload(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)
		assert.Equal(t, "", eth.GeneratePostPayload())
	})
}

func Test_GenerateSignPayload(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)

		result := eth.GenerateSignPayload()
		assert.Contains(t, result, "\"identity\":\""+strings.ToLower(crypto.PubkeyToAddress(wallet_sk.PublicKey).Hex()))
		assert.Contains(t, result, "\"persona\":\"0x"+mycrypto.CompressedPubkeyHex(eth.Pubkey))
		assert.Contains(t, result, "\"platform\":\"ethereum\"")
	})
}

func Test_Validate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)

		assert.Nil(t, eth.Validate())
	})
}
