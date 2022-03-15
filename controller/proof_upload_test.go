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
			ProofLocation: "1503630530465599488",
			PublicKey:     "0x037b721d6d84b474edbdab4d0746e9c777f60c414f9b0e651dd08272cb30ed6232",
			CreatedAt:     "1647327932",
			Uuid:          "ed9f421d-92e1-4c80-9bff-8516ef46ff43",
		}
		resp := ErrorResponse{}
		APITestCall(Engine, "POST", "/v1/proof", &req, &resp)
		assert.Empty(t, resp.Message)

		pc := model.ProofChain{
			Action:   req.Action,
			Platform: req.Platform,
			Identity: req.Identity,
			Location: req.ProofLocation,
		}
		model.DB.Where(&pc).First(&pc)
		assert.Greater(t, pc.ID, int64(0))
		assert.Equal(t, req.PublicKey, pc.Persona)

		proof := model.Proof{
			Platform: req.Platform,
			Identity: req.Identity,
			Location: req.ProofLocation,
		}
		model.DB.Where(&proof).First(&proof)
		assert.Greater(t, proof.ID, int64(0))
		assert.Equal(t, req.PublicKey, proof.Persona)
	})
}
