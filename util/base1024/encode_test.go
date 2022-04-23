package base1024

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEncodeToString(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		res := EncodeToString([]byte("Maskbook"))
		assert.Equal(t, "🐟🔂🏁🤖💧🚊😤", res)

	})

	t.Run("fail", func(t *testing.T) {
		res := EncodeToString([]byte("MaskBook"))
		assert.NotEqual(t, "🐟🔂🏁🤖💧🚊😤", res)

	})
}
