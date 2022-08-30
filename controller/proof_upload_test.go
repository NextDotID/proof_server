package controller

import (
	"testing"

	"github.com/nextdotid/proof_server/model"
	"github.com/nextdotid/proof_server/types"
	"github.com/nextdotid/proof_server/util"
	"github.com/stretchr/testify/assert"
)

func Test_ProofUpload(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)

		req := ProofUploadRequest{
			Action:        types.Actions.Create,
			Platform:      types.Platforms.Twitter,
			Identity:      "yeiwb",
			ProofLocation: "1504363098328924163",
			PublicKey:     "0x03666b700aeb6a6429f13cbb263e1bc566cd975a118b61bc796204109c1b351d19",
			CreatedAt:     "1647503071",
			Uuid:          "c6fa1483-1bad-4f07-b661-678b191ab4b3",
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
		orig_created_at, _ := util.TimestampStringToTime(req.CreatedAt)
		assert.Equal(t, pc.CreatedAt, orig_created_at)

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
