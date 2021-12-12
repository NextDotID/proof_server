package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_proofPayload(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)
		req := ProofPayloadRequest{
			Action:    "create",
			Platform:  "twitter",
			Identity:  "yeiwb",
			PublicKey: "0x028c3cda474361179d653c41a62f6bbb07265d535121e19aedf660da2924d0b1e3",
		}
		resp := ProofPayloadResponse{}
		APITestCall(Engine, "POST", "/v1/proof/payload", &req, &resp)
		assert.Contains(t, resp.SignPayload, "\"action\":\"create\"")
		assert.Contains(t, resp.SignPayload, "\"platform\":\"twitter\"")
		assert.Contains(t, resp.SignPayload, "\"identity\":\"yeiwb\"")
		assert.Contains(t, resp.SignPayload, "\"prev\":null")

		assert.Contains(t, resp.PostContent, "Prove myself:")
		assert.Contains(t, resp.PostContent, req.PublicKey)
		assert.Contains(t, resp.PostContent, "Signature:")
		assert.Contains(t, resp.PostContent, "%SIG_BASE64%")
	})
}
