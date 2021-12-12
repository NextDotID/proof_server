package controller

import (
	"testing"

	"github.com/nextdotid/proof-server/model"
	"github.com/nextdotid/proof-server/types"
	"github.com/stretchr/testify/assert"
)

func Test_ProofUpload(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)

		req := ProofUploadRequest{
			Action:        types.Actions.Create,
			Platform:      types.Platforms.Twitter,
			Identity:      "yeiwb",
			ProofLocation: "1469221200140574721",
			PublicKey:     "0x028c3cda474361179d653c41a62f6bbb07265d535121e19aedf660da2924d0b1e3",
		}
		resp := ErrorResponse{}
		APITestCall(Engine, "POST", "/v1/proof", &req, &resp)
		assert.Empty(t, resp.Message)

		proof := model.Proof{
			Platform: req.Platform,
			Identity: req.Identity,
			Location: req.ProofLocation,
		}
		model.DB.Where(&proof).First(&proof)
		assert.Greater(t, proof.ID, uint(0))
		assert.Equal(t, req.PublicKey, proof.Persona)
	})
}
