package controller

import (
	"encoding/base64"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nextdotid/proof-server/model"
	"github.com/nextdotid/proof-server/util/crypto"
	"github.com/stretchr/testify/assert"
)

var (
	kv_payload = model.KVPatch{
		Set: map[string]interface{}{
			"this": "is",
			"a": []string{"test", "case"},
		},
		Del: []string{"non.exist"},
	}
)

func test_kv_patch_generate_request() (KVPatchRequest) {
	pk, sk := crypto.GenerateKeypair()

	payload_resp := KVPayloadResponse{}
	APITestCall(
		Engine,
		"POST",
		"/v1/kv/payload",
		KVPayloadRequest{
			Persona: crypto.CompressedPubkeyHex(pk),
			Changes: kv_payload,
		},
		&payload_resp,
	)
	signature, _ := crypto.SignPersonal([]byte(payload_resp.SignPayload), sk)
	sig_base64 := base64.StdEncoding.EncodeToString(signature)

	req := KVPatchRequest{
		Persona:         crypto.CompressedPubkeyHex(pk),
		SignatureBase64: sig_base64,
		Changes:         kv_payload,
	}
	return req
}

func Test_kvPatch(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)
		req := test_kv_patch_generate_request()
		resp := gin.H{}

		resp_raw := APITestCall(Engine, "POST", "/v1/kv", req, &resp)
		assert.Equal(t, 201, resp_raw.Code)
		assert.Equal(t, []byte("{}"), resp_raw.Body.Bytes())

		pc, err := model.ProofChainFindLatest(req.Persona)
		assert.Nil(t, err)
		t.Logf("%+s", pc.Extra.String())
		assert.Equal(t, []string{"non.exist"}, pc.UnmarshalExtra().KVPatch.Del)

		kv, err := model.KVFindByPersona(model.MarshalPersona(req.Persona))
		assert.Nil(t, err)
		assert.NotNil(t, kv)

		content, err := kv.GetContent()
		assert.Nil(t, err)
		assert.Equal(t, "is", content["this"])
		assert.Equal(t, []interface{}{"test", "case"}, content["a"])
	})
}
