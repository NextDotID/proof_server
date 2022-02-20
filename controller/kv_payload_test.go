package controller

import (
	"testing"

	"github.com/nextdotid/proof-server/model"
	"github.com/stretchr/testify/assert"
)

const (
	pubkey_hex = "0x028c3cda474361179d653c41a62f6bbb07265d535121e19aedf660da2924d0b1e3"
)

func Test_kvPatchPayload(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)

		req := KVPayloadRequest{
			Persona: pubkey_hex,
			Changes: model.KVPatch{
				Set: map[string]interface{}{
					"this": "is",
					"a": "test",
				},
				Del: []string{},
			},
		}
		resp := KVPayloadResponse{}
		result := APITestCall(Engine, "POST", "/v1/kv/payload", req, &resp)
		assert.Equal(t, 200, result.Code)
		assert.Contains(t, resp.SignPayload, "kv")
		assert.Contains(t, resp.SignPayload, "\"set\"")
		assert.Contains(t, resp.SignPayload, "\"del\"")
		assert.Contains(t, resp.SignPayload, "\"prev\":null")
		assert.Contains(t, resp.SignPayload, "\"this\":\"is\"")
	})
}
