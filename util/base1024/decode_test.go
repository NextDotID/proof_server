package base1024

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDecodeString(t *testing.T) {
	t.Run("Equally", func(t *testing.T) {
		str := "🐟🔂🏁🤖💧🚊😤"
		res, err := DecodeString(str)
		assert.Nil(t, err)
		assert.Equal(t, "Maskbook", string(res))
	})

	t.Run("Not Equal", func(t *testing.T) {
		str := "🐟🔂🏁🤖💧"
		res, err := DecodeString(str)
		assert.Nil(t, err)
		assert.NotEqual(t, "Maskbook", string(res))
	})

	t.Run("Decode Playload", func(t *testing.T) {
		str := "👲🍚🍾🔆🏠🚱👧🦢🕟🛷🔭💘😝🙂🚳🦜🔙🔊🚗🏏👪🛹🗣🏳🦐🥫🦺🚎🕗🚷💡🚁🎟🗯📰🐊🕳🥠💐🎛🤵🆘🔣📥🦝🔉🌊🥠🥅🍏🥜🃏"
		res, err := DecodeString(str)
		assert.Nil(t, err)
		assert.NotEqual(t, "Maskbook", string(res))
	})
}
