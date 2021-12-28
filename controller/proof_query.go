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
	Persona string                          `json:"persona"`
	Proofs  []ProofQueryResponseSingleProof `json:"proofs"`
}

type ProofQueryResponseSingleProof struct {
	Platform types.Platform `json:"platform"`
	Identity string         `json:"identity"`
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

	proofs := make([]model.Proof, 0, 0)
	if req.Platform == types.Platforms.NextID {

		tx := model.DB.Where("persona", req.Identity).Find(&proofs)
		if tx.Error != nil || tx.RowsAffected == int64(0) || len(proofs) == 0 {
			return result
		}
	} else {
		tx := model.DB.
			Where("platform", req.Platform).
			Where("identity LIKE ?", "%"+req.Identity+"%").
			Find(&proofs)
		if tx.Error != nil || tx.RowsAffected == int64(0) || len(proofs) == 0 {
			return result
		}
	}

	// proofs.group_by(&:persona)
	persona_proof_map := make(map[string][]*model.Proof, 0)
	for _, p := range proofs {
		persona_proof, ok := persona_proof_map[p.Persona]
		if ok {
			persona_proof = append(persona_proof, &p)
		} else {
			persona_proof_map[p.Persona] = append(make([]*model.Proof, 0, 0), &p)
		}
	}

	for persona, proofs := range persona_proof_map {
		single := ProofQueryResponseSingle{
			Persona: persona,
			Proofs:  make([]ProofQueryResponseSingleProof, 0),
		}
		for _, p := range proofs {
			single.Proofs = append(single.Proofs, ProofQueryResponseSingleProof{
				Platform: p.Platform,
				Identity: p.Identity,
			})
		}
		result = append(result, single)
	}

	return result
}
