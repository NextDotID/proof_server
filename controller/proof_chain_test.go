package controller

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)


func Test_proofChainQuery(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)
		insert_proof(t)
		resp_body := ProofChainResponse{}
		resp := APITestCall(
			Engine,
			"GET",
			fmt.Sprintf("/v1/proofchain?public_key=%s", persona),
			nil,
			&resp_body,
		)
		assert.Equal(t, 200, resp.Code)
		t.Logf("%s", resp.Body.String())
		assert.Greater(t, len(resp_body.ProofChains), 0)
	})
}
