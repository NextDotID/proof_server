package controller

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nextdotid/proof_server/model"
	"github.com/samber/lo"
)

type ProofChainSingleRequest struct {
	LastID int `form:"last_id"`
	Count  int `form:"count"`
}

type ProofChainSingleResponse struct {
	Links []ProofChainSingleItem `json:"links"`
}

type ProofChainSingleItem struct {
	model.ProofChainItem
	Avatar string `json:"avatar"`
	ID     int64  `json:"id"`
}

// proofChainChanges returns proof chain one by one, unlinked.
func proofChainChanges(c *gin.Context) {
	req := ProofChainSingleRequest{}
	if err := c.BindQuery(&req); err != nil {
		errorResp(c, http.StatusBadRequest, errors.New("Param error"))
		return
	}
	if req.Count <= 0 {
		req.Count = 10
	}
	if req.Count >= 100 {
		req.Count = 100
	}

	pc_found := make([]model.ProofChain, 0, 0)
	tx := model.ReadOnlyDB.Where("id > ?", req.LastID).Limit(req.Count).Order("id ASC").Find(&pc_found)
	if tx.Error != nil {
		errorResp(c, http.StatusInternalServerError, tx.Error)
		return
	}

	proof_chains := lo.Map(pc_found, func(pc model.ProofChain, i int) ProofChainSingleItem {
		return ProofChainSingleItem{
			ProofChainItem: pc.ToProofChainItem(),
			Avatar:         pc.Persona,
			ID:             pc.ID,
		}
	})

	c.JSON(http.StatusOK, ProofChainSingleResponse{
		Links: proof_chains,
	})
}
