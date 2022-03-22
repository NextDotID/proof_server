package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/nextdotid/proof-server/model"
	"github.com/nextdotid/proof-server/types"
	"golang.org/x/xerrors"
	"net/http"
	"strconv"
)

type ProofChainRequest struct {
	PersonaPubkeyHex string `form:"public_key"`
	Page             int    `form:"page"`
}

type ProofChainResponse struct {
	Pagination  ProofChainPaginationResponse `json:"pagination"`
	ProofChains []ProofItem                  `form:"links"`
}

type ProofItem struct {
	ID            int64          `json:"id"`
	Prev          int64          `json:"prev"`
	Action        types.Action   `json:"action"`
	Platform      types.Platform `json:"platform"`
	Identity      string         `json:"identity"`
	ProofLocation string         `json:"proof_location"`
	CreatedAt     string         `json:"created_at"`
	Signature     string         `json:"signature"`
}

type ProofChainPaginationResponse struct {
	Total   int64 `json:"total"`
	Per     int   `json:"per"`
	Current int   `json:"current"`
	Next    int   `json:"next"`
}

func proofChainQuery(c *gin.Context) {
	req := ProofChainRequest{}
	if err := c.BindQuery(&req); err != nil {
		errorResp(c, http.StatusBadRequest, xerrors.Errorf("Param error"))
		return
	}
	if len(req.PersonaPubkeyHex) == 0 {
		errorResp(c, http.StatusBadRequest, xerrors.Errorf("Param missing"))
		return
	}

	list, pagination := performProofChainQuery(req)
	c.JSON(http.StatusOK, ProofChainResponse{
		Pagination:  pagination,
		ProofChains: list,
	})

}

func performProofChainQuery(req ProofChainRequest) ([]ProofItem, ProofChainPaginationResponse) {
	pagination := ProofChainPaginationResponse{
		Total:   0,
		Per:     PER_PAGE,
		Current: req.Page,
		Next:    0,
	}
	if pagination.Current <= 0 { // `page` param not provided. Set it to 1.
		pagination.Current = 1
	}
	offsetCount := pagination.Per * (pagination.Current - 1)

	rs := make([]ProofItem, 0, 0)
	proofs := make([]model.ProofChain, 0, 0)
	tx := model.DB.Model(&model.ProofChain{})
	tx = tx.Where("persona = ?", req.PersonaPubkeyHex)

	countTx := tx // Value-copy another query for total amount calculation
	countTx.Count(&pagination.Total)

	if pagination.Total > int64(pagination.Per*pagination.Current) {
		pagination.Next = pagination.Current + 1
	}
	tx = tx.Offset(offsetCount).Limit(pagination.Per).Find(&proofs)

	if tx.Error != nil || tx.RowsAffected == int64(0) || len(proofs) == 0 {
		return rs, pagination
	}
	for _, item := range proofs {
		rs = append(rs, ProofItem{
			ID:            item.ID,
			Prev:          item.PreviousID.Int64,
			Action:        item.Action,
			Platform:      item.Platform,
			Identity:      item.Identity,
			ProofLocation: item.Location,
			CreatedAt:     strconv.FormatInt(item.CreatedAt.Unix(), 10),
			Signature:     item.Signature,
		})
	}
	return rs, pagination
}
