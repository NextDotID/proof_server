package ethereum

import (
	"crypto/ecdsa"
	"encoding/base64"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	"github.com/nextdotid/proof-server/config"
	"github.com/nextdotid/proof-server/types"
	"github.com/nextdotid/proof-server/validator"

	mycrypto "github.com/nextdotid/proof-server/util/crypto"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var (
	persona_sk *ecdsa.PrivateKey
	wallet_sk  *ecdsa.PrivateKey
)

func before_each(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	config.Init("../../config/config.test.json")
}

func generate() Ethereum {
	eth := Ethereum{
		Base: &validator.Base{
			Platform: types.Platforms.Ethereum,
			Previous: "",
			Action:   types.Actions.Create,
			Extra: map[string]string{
				"wallet_signature": "",
			},
			CreatedAt: time.Now(),
			Uuid:      uuid.New(),
		},
	}
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

	return eth
}

func Test_GeneratePostPayload(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)
		eth := generate()
		assert.Equal(t, "", eth.GeneratePostPayload()["default"])
	})
}

func Test_GenerateSignPayload(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)

		eth := generate()
		result := eth.GenerateSignPayload()
		assert.Contains(t, result, "\"identity\":\""+strings.ToLower(crypto.PubkeyToAddress(wallet_sk.PublicKey).Hex()))
		assert.Contains(t, result, "\"persona\":\"0x"+mycrypto.CompressedPubkeyHex(eth.Pubkey))
		assert.Contains(t, result, "\"platform\":\"ethereum\"")
	})
}

func Test_Validate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)

		eth := generate()
		assert.Nil(t, eth.Validate())
	})
}

func Test_Validate_Delete(t *testing.T) {
	t.Run("signed by persona", func(t *testing.T) {
		before_each(t)

		eth := generate()
		eth.Action = types.Actions.Delete
		eth.Extra = map[string]string{
			"wallet_signature": "",
		}
		eth.Signature, _ = mycrypto.SignPersonal([]byte(eth.GenerateSignPayload()), persona_sk)

		assert.Nil(t, eth.Validate())
	})

	t.Run("signed by wallet", func(t *testing.T) {
		before_each(t)

		eth := generate()
		eth.Action = types.Actions.Delete
		wallet_sig, _ := mycrypto.SignPersonal([]byte(eth.GenerateSignPayload()), wallet_sk)
		eth.Extra = map[string]string{
			"wallet_signature": base64.StdEncoding.EncodeToString(wallet_sig),
		}

		assert.Nil(t, eth.Validate())
	})

	t.Run("signed by persona, but put in wallet_signature", func(t *testing.T) {
		before_each(t)

		eth := generate()
		eth.Action = types.Actions.Delete

		eth.Signature, _ = mycrypto.SignPersonal([]byte(eth.GenerateSignPayload()), persona_sk)
		eth.Extra = map[string]string{
			"wallet_signature": base64.StdEncoding.EncodeToString(eth.Signature),
		}

		assert.NotNil(t, eth.Validate())
	})

	t.Run("signed by wallet, but put in eth.Signature", func(t *testing.T) {
		before_each(t)

		before_each(t)

		eth := generate()
		eth.Action = types.Actions.Delete
		eth.Signature, _ = mycrypto.SignPersonal([]byte(eth.GenerateSignPayload()), wallet_sk)
		eth.Extra = map[string]string{}

		assert.NotNil(t, eth.Validate())
	})
}
