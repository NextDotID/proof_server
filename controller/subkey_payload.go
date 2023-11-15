package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/nextdotid/proof_server/model"
	"github.com/nextdotid/proof_server/types"
)

type subkeyPayloadRequest struct {
	Avatar    string                `json:"avatar"`
	Algorithm types.SubkeyAlgorithm `json:"algorithm"`
	PublicKey string                `json:"public_key"`
	RP_ID     string                `json:"rp_id"`
}

type subkeyPayloadResponse struct {
	SignPayload string `json:"sign_payload"`
}

// POST /v1/subkey/payload
func subkeyPayload(c *gin.Context) {
	req := subkeyPayloadRequest{}
	if err := c.BindJSON(&req); err != nil {
		errorResp(c, 400, err)
		return
	}

	subkey := model.Subkey{
		Algorithm: req.Algorithm,
		Avatar:    req.Avatar,
		PublicKey: req.PublicKey,
		RP_ID:     req.RP_ID,
	}
	payload, err := subkey.SignPayload()
	if err != nil {
		errorResp(c, 400, err)
		return
	}
	c.JSON(200, subkeyPayloadResponse{
		SignPayload: payload,
	})
}
