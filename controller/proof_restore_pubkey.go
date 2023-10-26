package controller

import (
	"encoding/base64"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nextdotid/proof_server/types"
	"github.com/nextdotid/proof_server/util"
	"github.com/nextdotid/proof_server/util/crypto"
	"github.com/nextdotid/proof_server/validator"
	"golang.org/x/xerrors"
)

type ProofRestorePubkeyRequest struct {
	Action    types.Action   `json:"action"`
	Platform  types.Platform `json:"platform"`
	Identity  string         `json:"identity"`
	Uuid      string         `json:"uuid"`
	CreatedAt string         `json:"created_at"`
	Previous  string         `json:"previous"`
	Signature string         `json:"signature"`
}

type ProofRestorePubkeyResponse struct {
	PublicKey string `json:"public_key"`
}


func proofRestorePubkey(c *gin.Context) {
	req := ProofRestorePubkeyRequest{};
	if err := c.BindJSON(&req); err != nil {
		errorResp(c, http.StatusBadRequest, xerrors.Errorf("Param type error"))
		return
	}
	if err := proofRestorePubkeyCheckRequest(&req); err != nil {
		errorResp(c, http.StatusBadRequest, err)
		return
	}
	createdAt, _ := util.TimestampStringToTime(req.CreatedAt)
	uuid, _ := uuid.Parse(req.Uuid)
	signature, _ := base64.StdEncoding.DecodeString(req.Signature)
	base := validator.Base{
		Platform:         req.Platform,
		Previous:         req.Previous,
		Action:           req.Action,
		Identity:         req.Identity,
		CreatedAt:        createdAt,
		Uuid:             uuid,
	}
	baseValidator := validator.BaseToInterface(&base)
	signPayload := baseValidator.GenerateSignPayload()
	pubkey, err := crypto.RecoverPubkeyFromPersonalSignature(signPayload, signature)
	if err != nil {
		errorResp(c, http.StatusBadRequest, xerrors.Errorf("restoring pubkey from sig: %w", err))
		return
	}
	c.JSON(http.StatusOK, ProofRestorePubkeyResponse{
		PublicKey: "0x" + crypto.CompressedPubkeyHex(pubkey),
	})
}

func proofRestorePubkeyCheckRequest(req *ProofRestorePubkeyRequest) error {
	if req.Platform == "" || req.Identity == "" || req.Action == "" || req.Uuid == "" || req.CreatedAt == "" || req.Signature == "" {
		return xerrors.Errorf("param missing")
	}
	_, err := uuid.Parse(req.Uuid)
	if err != nil {
		return xerrors.Errorf("UUID parse error: %w", err)
	}

	_, err = util.TimestampStringToTime(req.CreatedAt)
	if err != nil {
		return xerrors.Errorf("created_at parse error: %w", err)
	}

	signature, err := base64.StdEncoding.DecodeString(req.Signature)
	if err != nil {
		return xerrors.Errorf("signature parse error: %w", err)
	}
	if len(signature) != 65 {
		return xerrors.Errorf("signature length error: expect 65, got %d", len(signature))
	}

	if len(req.Previous) > 0{
		previous_signature, err := base64.StdEncoding.DecodeString(req.Previous)
		if err != nil {
			return xerrors.Errorf("previous signature parse error: %w", err)
		}
		if len(previous_signature) != 65 {
			return xerrors.Errorf("previous signature length error: expect 65, got %d", len(previous_signature))
		}
	}

	return nil
}
