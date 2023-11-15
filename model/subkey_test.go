package model

import (
	"crypto/ecdsa"
	"testing"

	"github.com/nextdotid/proof_server/types"
	"github.com/nextdotid/proof_server/util/crypto"
	"github.com/stretchr/testify/require"
)

func generateK1Subkey() (subkey *Subkey, avatarSK *ecdsa.PrivateKey, subkeySK *ecdsa.PrivateKey) {
	avatarPK, avatarSK := crypto.GenerateSecp256k1Keypair()
	subkeyPK, subkeySK := crypto.GenerateSecp256k1Keypair()
	avatarPKHex := "0x" + crypto.CompressedPubkeyHex(avatarPK)
	subkeyPKHex := "0x" + crypto.CompressedPubkeyHex(subkeyPK)
	return &Subkey{
		Name:      "Yubikey",
		RP_ID:     "apple.com",
		Avatar:    avatarPKHex,
		Algorithm: types.SubkeyAlgorithms.Secp256K1,
		PublicKey: subkeyPKHex,
	}, avatarSK, subkeySK
}

func Test_SignPayload(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		subkey, _, _ := generateK1Subkey()
		payload, err := subkey.SignPayload()
		require.NoError(t, err)
		require.Contains(t, payload, subkey.Avatar)
		require.Contains(t, payload, subkey.Algorithm)
	})
}

func Test_ValidateSignature(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		subkey, avatarSK, _ := generateK1Subkey()
		payload, _ := subkey.SignPayload()
		signature, err := crypto.SignPersonal([]byte(payload), avatarSK)
		require.NoError(t, err)
		// require.NoError(t, crypto.ValidatePersonalSignature(payload, signature, &avatarSK.PublicKey))
		require.NoError(t, subkey.ValidateSignature(payload, signature))
	})
}
