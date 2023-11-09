package controller

import (
	"encoding/base64"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nextdotid/proof_server/model"
	"github.com/nextdotid/proof_server/types"
	"golang.org/x/xerrors"
)

type subkeySubmitRequest struct {
	Avatar    string                `json:"avatar"`
	Algorithm types.SubkeyAlgorithm `json:"algorithm"`
	PublicKey string                `json:"public_key"`
	RP_ID     string                `json:"rp_id"`
	Name      string                `json:"name"`
	Signature string                `json:"signature"`
}

type subkeySubmitResponse struct {
	SignPayload string `json:"sign_payload"`
}

// POST /v1/subkey
func subkeySubmit(c *gin.Context) {
	req := subkeySubmitRequest{}
	if err := c.BindJSON(&req); err != nil {
		errorResp(c, 400, err)
		return
	}

	subkey := model.Subkey{
		CreatedAt: time.Now(),
		Name:      req.Name,
		RP_ID:     req.RP_ID,
		Avatar:    req.Avatar,
		Algorithm: req.Algorithm,
		PublicKey: req.PublicKey,
	}
	payload, err := subkey.SignPayload()
	if err != nil {
		errorResp(c, 400, err)
		return
	}
	signature, err := base64.StdEncoding.DecodeString(req.Signature)
	if err != nil {
		errorResp(c, 400, xerrors.Errorf("Error when decoding signature: %w", err))
		return
	}
	if err := subkey.ValidateSignature(payload, signature); err != nil {
		errorResp(c, 400, xerrors.Errorf("Error when validating signature: %w", err))
		return
	}

	tx := model.DB.Create(&subkey)
	if tx.Error != nil {
		errorResp(c, 500, xerrors.Errorf("Error when saving subkey: %w", err))
		return
	}

	c.JSON(200, subkeyPayloadResponse{
		SignPayload: payload,
	})
}
