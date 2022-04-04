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
		//t.Logf("%s", resp.Body.String())
		assert.Equal(t, 2, len(resp_body.ProofChains))
	})

	t.Run("empty result", func(t *testing.T) {
		before_each(t)
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

	t.Run("patination", func(t *testing.T) {
		before_each(t)
		for i := 0; i < 22; i++ { // Create 44 records
			insert_proof(t)
		}
		url := fmt.Sprintf("/v1/proofchain?public_key=%s", persona)

		resp_page1 := ProofChainResponse{} // Page not given
		APITestCall(Engine, "GET", url, nil, &resp_page1)
		assert.Equal(t, int64(44), resp_page1.Pagination.Total)
		assert.Equal(t, 1, resp_page1.Pagination.Current)
		assert.Equal(t, 2, resp_page1.Pagination.Next)
		assert.Equal(t, PER_PAGE, len(resp_page1.ProofChains))

		resp_page3 := ProofChainResponse{} // Last page
		APITestCall(Engine, "GET", url+"&page=3", nil, &resp_page3)
		assert.Equal(t, 3, resp_page3.Pagination.Current)
		assert.Equal(t, 0, resp_page3.Pagination.Next)
		assert.Equal(t, 4, len(resp_page3.ProofChains))

		resp_page4 := ProofChainResponse{} // Page overflow
		APITestCall(Engine, "GET", url+"&page=4", nil, &resp_page4)
		assert.Equal(t, 4, resp_page4.Pagination.Current)
		assert.Equal(t, 0, resp_page4.Pagination.Next)
		assert.Equal(t, 0, len(resp_page4.ProofChains))
	})
}
