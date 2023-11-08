package controller

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nextdotid/proof_server/model"
	"github.com/nextdotid/proof_server/util/crypto"
	"golang.org/x/xerrors"
)

type ProofExistsRequest struct {
	Platform         string `form:"platform"`
	Identity         string `form:"identity"`
	PersonaPubkeyHex string `form:"public_key"`
}

type ProofExistsResponse struct {
	CreatedAt     string `json:"created_at"`
	LastCheckedAt string `json:"last_checked_at"`
	IsValid       bool   `json:"is_valid"`
	InvalidReason string `json:"invalid_reason"`
}

func proofExists(c *gin.Context) {
	req := ProofExistsRequest{}
	if err := c.BindQuery(&req); err != nil {
		errorResp(c, http.StatusBadRequest, xerrors.Errorf("Param error"))
		return
	}
	if !proofExistsCheckRequest(&req) {
		errorResp(c, http.StatusBadRequest, xerrors.Errorf("Param missing"))
		return
	}

	personaPubkey, err := crypto.StringToSecp256k1Pubkey(req.PersonaPubkeyHex)
	if err != nil {
		errorResp(c, http.StatusBadRequest, xerrors.Errorf("Public key unmarshal error"))
		return
	}
	found := model.Proof{}
	tx := model.ReadOnlyDB.Where(
		"persona = ? AND platform = ? AND (identity = ? OR alt_id = ?)",
		model.MarshalAvatar(personaPubkey),
		req.Platform,
		strings.ToLower(req.Identity),
		strings.ToLower(req.Identity),
	).Find(&found)

	if tx.Error != nil {
		errorResp(c, http.StatusInternalServerError, xerrors.Errorf("Error in DB: %w", err))
		return
	}
	if found.ID == int64(0) { // Not found
		errorResp(c, http.StatusNotFound, xerrors.Errorf("Record not found for %s: %s", req.Platform, req.Identity))
		return
	}
	if found.IsOutdated() {
		go triggerRevalidate(found.ID)
	}

	c.JSON(http.StatusOK, ProofExistsResponse{
		CreatedAt:     strconv.FormatInt(found.CreatedAt.Unix(), 10),
		LastCheckedAt: strconv.FormatInt(found.LastCheckedAt.Unix(), 10),
		IsValid:       found.IsValid,
		InvalidReason: found.InvalidReason,
	})
}

func proofExistsCheckRequest(req *ProofExistsRequest) bool {
	return req.Identity != "" && req.Platform != "" && req.PersonaPubkeyHex != ""
}
