package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nextdotid/proof_server/model"
	"golang.org/x/xerrors"
)

type ProofChainRequest struct {
	Avatar string `form:"avatar"`
	Page   int    `form:"page"`
}

type ProofChainResponse struct {
	Pagination  ProofChainPaginationResponse `json:"pagination"`
	ProofChains []model.ProofChainItem       `json:"proof_chain"`
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
	if len(req.Avatar) == 0 {
		errorResp(c, http.StatusBadRequest, xerrors.Errorf("Param missing"))
		return
	}

	list, pagination, err := performProofChainQuery(req)
	if err != nil {
		errorResp(c, http.StatusInternalServerError, xerrors.Errorf("Error in DB: %w", err))
		return
	}

	c.JSON(http.StatusOK, ProofChainResponse{
		Pagination:  pagination,
		ProofChains: list,
	})
}

func performProofChainQuery(req ProofChainRequest) ([]model.ProofChainItem, ProofChainPaginationResponse, error) {
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

	total, rs, err := model.ProofChainFindByPersona(req.Avatar, false, offsetCount, pagination.Per)
	pagination.Total = total
	if total > int64(pagination.Per*pagination.Current) {
		pagination.Next = pagination.Current + 1
	}

	return rs, pagination, err
}
