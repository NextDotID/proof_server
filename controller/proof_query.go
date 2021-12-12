package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nextdotid/proof-server/model"
	"github.com/nextdotid/proof-server/types"
	"golang.org/x/xerrors"
)

type ProofQueryRequest struct {
	Platform types.Platform `form:"platform"`
	Identity string         `form:"identity"`
}

type ProofQueryResponse struct {
	IDs []ProofQueryResponseSingle `json:"ids"`
}

type ProofQueryResponseSingle struct {
	Platform      types.Platform `json:"platform"`
	Identity      string         `json:"identity"`
	ProofLocation string         `json:"proof_location"`
}

func proofQuery(c *gin.Context) {
	req := ProofQueryRequest{}
	if err := c.BindQuery(&req); err != nil {
		errorResp(c, http.StatusBadRequest, xerrors.Errorf("Param error"))
		return
	}
	if req.Platform == "" || req.Identity == "" {
		errorResp(c, http.StatusBadRequest, xerrors.Errorf("Param missing"))
		return
	}

	c.JSON(http.StatusOK, ProofQueryResponse{
		IDs: performProofQuery(req),
	})
}

func performProofQuery(req ProofQueryRequest) []ProofQueryResponseSingle {
	result := make([]ProofQueryResponseSingle, 0, 0)

	proof := &model.Proof{}
	// TODO: should deal with multi bindings
	tx := model.DB.Where(&model.Proof{Platform: req.Platform, Identity: req.Identity}).Last(proof)
	if tx.Error != nil || proof.ID == uint(0) {
		return result
	}

	proofs := make([]model.Proof, 0, 0)
	tx = model.DB.Where(&model.Proof{Persona: proof.Persona}).Find(&proofs)
	if tx.Error != nil || len(proofs) == 0 {
		return result
	}

	result = append(result, ProofQueryResponseSingle{
		Platform: types.Platforms.NextID,
		Identity: proofs[0].Persona,
	})

	for _, p := range proofs {
		result = append(result, ProofQueryResponseSingle{
			Platform: p.Platform,
			Identity: p.Identity,
		})
	}
	return result
}
