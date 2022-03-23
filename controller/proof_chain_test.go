package controller

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)


func Test_proofChainQuery(t *testing.T) {
	before_each(t)
	insert_proof(t)
	t.Run("success", func(t *testing.T) {
		resp_body := ProofChainResponse{}
		resp := APITestCall(
			Engine,
			"GET",
			fmt.Sprintf("/v1/proofchain?public_key=%s", persona),
			nil,
			&resp_body,
		)
		assert.Equal(t, 200, resp.Code)
		//t.Logf("%s", resp.Body.String())
		assert.Equal(t, 2, len(resp_body.ProofChains))
	})

	t.Run("failure", func(t *testing.T) {
		resp_body := ProofChainResponse{}
		resp := APITestCall(
			Engine,
			"GET",
			fmt.Sprintf("/v1/proofchain?public_key=%s", "aaa"),
			nil,
			&resp_body,
		)
		assert.Equal(t, 200, resp.Code)
		//t.Logf("%s", resp.Body.String())
		assert.Equal(t, 0, len(resp_body.ProofChains))
	})
}
