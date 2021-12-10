package crypto

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/stretchr/testify/assert"
)

func Test_SignVerify(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		payload := "test123"
		pk, sk := GenerateKeypair()
		signature, err := SignPersonal([]byte(payload), sk)
		assert.Nil(t, err)

		result := ValidatePersonalSignature(payload, signature, common.Bytes2Hex(crypto.FromECDSAPub(pk)))
		assert.True(t, result)
	})

	t.Run("fail if pubkey mismatch", func(t *testing.T) {
		payload := "test123"
		_, sk := GenerateKeypair()
		signature, _ := SignPersonal([]byte(payload), sk)

		new_pk, _ := GenerateKeypair()
		result := ValidatePersonalSignature(payload, signature, common.Bytes2Hex(crypto.FromECDSAPub(new_pk)))
		assert.False(t, result)
	})

	t.Run("fail if payload mismatch", func(t *testing.T) {
		payload := "test123"
		pk, sk := GenerateKeypair()
		signature, _ := SignPersonal([]byte(payload), sk)

		result := ValidatePersonalSignature("foobar", signature, common.Bytes2Hex(crypto.FromECDSAPub(pk)))
		assert.False(t, result)
	})

	t.Run("fail if signature mismatch", func(t *testing.T) {
		pk, sk := GenerateKeypair()
		signature, _ := SignPersonal([]byte("foobar"), sk)

		result := ValidatePersonalSignature("test123", signature, common.Bytes2Hex(crypto.FromECDSAPub(pk)))
		assert.False(t, result)
	})
}
