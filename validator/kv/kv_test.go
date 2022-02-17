package kv

import (
	"encoding/json"
	"testing"
	"github.com/stretchr/testify/assert"
)

const (
	test_pubkey = "0x03ec5ff37fe2e0b0ff4e934796f42450184fb0dbbda33ff436a0cf0632e6a4c499"
)

func Test_UnmarshalAnything(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		json_string := "{\"this\": \"is\", \"a\": [\"test\"]}"
		var target map[string]interface{}
		err := json.Unmarshal([]byte(json_string), &target)
		assert.Nil(t, err)

		marshaled, err := json.Marshal(target)
		assert.Nil(t, err)
		assert.Equal(t, []byte("{\"a\":[\"test\"],\"this\":\"is\"}"), marshaled)
	})
}
