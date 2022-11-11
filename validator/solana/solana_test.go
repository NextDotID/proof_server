package solana

import (
	"crypto/ecdsa"
	"strings"
	"testing"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/google/uuid"
	"github.com/mr-tron/base58"
	"github.com/nextdotid/proof_server/config"
	"github.com/nextdotid/proof_server/types"
	"github.com/nextdotid/proof_server/validator"

	mycrypto "github.com/nextdotid/proof_server/util/crypto"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

var (
	personaPriv *ecdsa.PrivateKey
	walletPriv  solana.PrivateKey
)

func before_each(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	config.Init("../../config/config.test.json")
}

func generate() Solana {
	sol := Solana{
		Base: &validator.Base{
			Platform: types.Platforms.Solana,
			Previous: "",
			Action:   types.Actions.Create,
			Extra: map[string]string{
				"wallet_signature": "",
			},
			CreatedAt: time.Now(),
			Uuid:      uuid.New(),
		},
	}
	_, personaPriv = mycrypto.GenerateKeypair()
	sol.Pubkey = &personaPriv.PublicKey

	walletPriv, _ = solana.NewRandomPrivateKey()

	// Base58 encoded
	sol.Identity = walletPriv.PublicKey().String()

	// Generate sig
	payloadBytes := []byte(sol.GenerateSignPayload())
	sol.Signature, _ = mycrypto.SignPersonal(payloadBytes, personaPriv)
	walletSig, _ := walletPriv.Sign(payloadBytes)
	sol.Extra = map[string]string{
		// Base58 encoded
		"wallet_signature": walletSig.String(),
	}

	return sol
}

func Test_GeneratePostPayload(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)
		sol := generate()
		require.Equal(t, "", sol.GeneratePostPayload()["default"])
	})
}

func Test_GenerateSignPayload(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)

		sol := generate()
		result := sol.GenerateSignPayload()
		require.Contains(t, result, "\"identity\":\""+walletPriv.PublicKey().String())
		require.NotContains(t, result, "\"identity\":\""+strings.ToLower(walletPriv.PublicKey().String()))
		require.Contains(t, result, "\"persona\":\"0x"+mycrypto.CompressedPubkeyHex(sol.Pubkey))
		require.Contains(t, result, "\"platform\":\"solana\"")
	})
}

func Test_Validate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)

		sol := generate()
		require.NoError(t, sol.Validate())
	})

	t.Run("fail with wrong wallet signature", func(t *testing.T) {
		before_each(t)

		sol := generate()
		sol.Extra = map[string]string{
			"wallet_signature": base58.Encode([]byte("{}")),
		}
		sol.Signature, _ = mycrypto.SignPersonal([]byte(sol.GenerateSignPayload()), personaPriv)

		require.Error(t, sol.Validate())
	})

	t.Run("fail with wrong persona signature", func(t *testing.T) {
		before_each(t)

		sol := generate()
		walletSig, _ := walletPriv.Sign([]byte(sol.GenerateSignPayload()))
		sol.Extra = map[string]string{
			"wallet_signature": walletSig.String(),
		}
		sol.Signature = []byte(uuid.New().String())

		require.Error(t, sol.Validate())
	})
}

func Test_Validate_Delete(t *testing.T) {
	t.Run("signed by persona", func(t *testing.T) {
		before_each(t)

		sol := generate()
		sol.Action = types.Actions.Delete
		sol.Extra = map[string]string{
			"wallet_signature": "",
		}
		sol.Signature, _ = mycrypto.SignPersonal([]byte(sol.GenerateSignPayload()), personaPriv)

		require.NoError(t, sol.Validate())
		require.Equal(t, sol.Identity, sol.AltID)
	})

	t.Run("signed by wallet", func(t *testing.T) {
		before_each(t)

		sol := generate()
		sol.Action = types.Actions.Delete
		walletSig, _ := walletPriv.Sign([]byte(sol.GenerateSignPayload()))
		sol.Extra = map[string]string{
			"wallet_signature": walletSig.String(),
		}

		require.NoError(t, sol.Validate())
	})

	t.Run("signed by persona, but with wrong wallet_signature", func(t *testing.T) {
		before_each(t)

		sol := generate()
		sol.Action = types.Actions.Delete
		sol.Signature, _ = mycrypto.SignPersonal([]byte(sol.GenerateSignPayload()), personaPriv)
		sol.Extra = map[string]string{
			"wallet_signature": base58.Encode([]byte(uuid.New().String())),
		}

		require.Error(t, sol.Validate())
	})

	t.Run("signed by wallet, but with wrong persona sig, which should be ok", func(t *testing.T) {
		before_each(t)

		sol := generate()
		sol.Action = types.Actions.Delete
		walletSig, _ := walletPriv.Sign([]byte(sol.GenerateSignPayload()))
		sol.Signature = []byte(uuid.New().String())
		sol.Extra = map[string]string{
			"wallet_signature": walletSig.String(),
		}

		require.NoError(t, sol.Validate())
	})
}
