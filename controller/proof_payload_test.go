package controller

import (
	"encoding/json"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nextdotid/proof_server/model"
	"github.com/nextdotid/proof_server/util/crypto"
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

		assert.Contains(t, resp.PostContent["default"], "Verifying my Twitter ID")
		assert.Contains(t, resp.PostContent["default"], req.Identity)
		assert.Contains(t, resp.PostContent["default"], "Sig:")
		assert.Contains(t, resp.PostContent["default"], "%SIG_BASE64%")

		assert.True(t, len(resp.Uuid) > 0)
		assert.True(t, len(resp.CreatedAt) > 0)
	})

	t.Run("with previous", func(t *testing.T) {
		before_each(t)
		pk, _ := crypto.StringToSecp256k1Pubkey("0x028c3cda474361179d653c41a62f6bbb07265d535121e19aedf660da2924d0b1e3")

		proof := model.ProofChain{
			Persona:   "0x" + crypto.CompressedPubkeyHex(pk),
			Platform:  "twitter",
			Identity:  "yeiwb",
			Location:  "1469221200140574721",
			Signature: "gMUJ75eewkdaNrFp7bafzckv9+rlW7rVaxkB7/sYzYgFdFltYG+gn0lYzVNgrAdHWZPmu2giwJniGG7HG9iNigE=",
		}
		tx := model.DB.Create(&proof)
		assert.Nil(t, tx.Error)

		req := ProofPayloadRequest{
			Action:    "delete",
			Platform:  "twitter",
			Identity:  "yeiwb",
			PublicKey: "0x" + crypto.CompressedPubkeyHex(pk),
		}
		resp := ProofPayloadResponse{}

		APITestCall(Engine, "POST", "/v1/proof/payload", &req, &resp)
		sign_payload := gin.H{}

		assert.Nil(t, json.Unmarshal([]byte(resp.SignPayload), &sign_payload))
		prev, ok := sign_payload["prev"]
		assert.True(t, ok)
		t.Logf("Prev: %s", prev)
		assert.Equal(t, prev, proof.Signature)
	})

}
