package crypto

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ToPubkey(t *testing.T) {
	t.Run("secp256k1", func(t *testing.T) {
		pk, _ := GenerateSecp256k1Keypair()
		compressed := "0x" + common.Bytes2Hex(crypto.CompressPubkey(pk))
		pkRecovered, err := StringToSecp256k1Pubkey(compressed)
		require.NoError(t, err)
		require.Equal(t, pk.X.String(), pkRecovered.X.String())
		require.Equal(t, pk.Y.String(), pkRecovered.Y.String())
	})
}

func Test_Secp256k1_SignVerify(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		payload := "test123"
		pk, sk := GenerateSecp256k1Keypair()
		signature, err := SignPersonal([]byte(payload), sk)
		assert.Nil(t, err)

		err = ValidatePersonalSignature(payload, signature, pk)
		assert.Nil(t, err)
	})

	t.Run("fail if pubkey mismatch", func(t *testing.T) {
		payload := "test123"
		_, sk := GenerateSecp256k1Keypair()
		signature, _ := SignPersonal([]byte(payload), sk)

		new_pk, _ := GenerateSecp256k1Keypair()
		err := ValidatePersonalSignature(payload, signature, new_pk)
		assert.NotNil(t, err)
	})

	t.Run("fail if payload mismatch", func(t *testing.T) {
		payload := "test123"
		pk, sk := GenerateSecp256k1Keypair()
		signature, _ := SignPersonal([]byte(payload), sk)

		err := ValidatePersonalSignature("foobar", signature, pk)
		assert.NotNil(t, err)
	})

	t.Run("fail if signature mismatch", func(t *testing.T) {
		pk, sk := GenerateSecp256k1Keypair()
		signature, _ := SignPersonal([]byte("foobar"), sk)

		err := ValidatePersonalSignature("test123", signature, pk)
		assert.NotNil(t, err)
	})
}
