package controller

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_proofExists(t *testing.T) {
	t.Run("smoke", func(t *testing.T) {
		before_each(t)

		resp := ErrorResponse{}
		APITestCall(
			Engine,
			"GET",
			fmt.Sprintf("/v1/proof/exists?platform=twitter&identity=test&public_key=%s", persona),
			nil,
			&resp,
		)
		assert.Contains(t, resp.Message, "not found")
	})

	t.Run("success", func(t *testing.T) {
		before_each(t)
		insert_proof(t)
		resp_body := ProofExistsResponse{}
		resp := APITestCall(
			Engine,
			"GET",
			fmt.Sprintf("/v1/proof/exists?platform=twitter&identity=yeiwb&public_key=%s", persona),
			nil,
			&resp_body,
		)
		assert.Equal(t, 200, resp.Code)
		t.Logf("%s", resp.Body.String())
		assert.Equal(t, true, resp_body.IsValid)
		assert.Equal(t, "", resp_body.InvalidReason)
	})
}
