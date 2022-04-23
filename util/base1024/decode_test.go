package base1024

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDecodeString(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		str := "🐟🔂🏁🤖💧🚊😤"
		res, err := DecodeString(str)
		assert.Nil(t, err)
		assert.Equal(t, "Maskbook", string(res))

	})
}
