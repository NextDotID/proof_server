package controller

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nextdotid/proof_server/model"
	"github.com/nextdotid/proof_server/types"
	"github.com/nextdotid/proof_server/util"
	"github.com/nextdotid/proof_server/util/crypto"
	"github.com/nextdotid/proof_server/validator"
	"golang.org/x/xerrors"
)

type ProofPayloadRequest struct {
	Action    types.Action             `json:"action"`
	Platform  types.Platform           `json:"platform"`
	Identity  string                   `json:"identity"`
	PublicKey string                   `json:"public_key"`
	Extra     ProofPayloadRequestExtra `json:"extra"`
}

type ProofPayloadResponse struct {
	PostContent map[string]string `json:"post_content"`
	SignPayload string            `json:"sign_payload"`
	Uuid        string            `json:"uuid"`
	CreatedAt   string            `json:"created_at"`
}

type ProofPayloadRequestExtra struct {
	EthereumWalletSignature string `json:"wallet_signature"`
}

func proofPayload(c *gin.Context) {
	req := &ProofPayloadRequest{}
	err := c.BindJSON(req)
	if err != nil {
		errorResp(c, http.StatusBadRequest, xerrors.Errorf("when parsing body: %w", err))
		return
	}
	if !proofPayloadCheckRequest(req) {
		errorResp(c, http.StatusBadRequest, xerrors.New("param invalid"))
		return
	}

	parsed_pubkey, err := crypto.StringToSecp256k1Pubkey(req.PublicKey)
	if err != nil {
		errorResp(c, http.StatusBadRequest, xerrors.New("public key not recognized"))
		return
	}

	previous_pc, err := model.ProofChainFindLatest(crypto.CompressedPubkeyHex(parsed_pubkey))
	if err != nil {
		errorResp(c, http.StatusInternalServerError, xerrors.New("previous proof not found"))
		return
	}

	var previous_signature string
	if previous_pc == nil {
		previous_signature = ""
	} else {
		previous_signature = previous_pc.Signature
	}

	v := validator.Base{
		Platform:  req.Platform,
		Previous:  previous_signature,
		Action:    req.Action,
		Pubkey:    parsed_pubkey,
		Identity:  req.Identity,
		Uuid:      uuid.New(),
		CreatedAt: time.Now(),
		Extra: map[string]string{
			"wallet_signature": req.Extra.EthereumWalletSignature,
		},
	}

	performer := validator.BaseToInterface(&v)
	if performer == nil {
		errorResp(c, http.StatusBadRequest, xerrors.New("unknown platform"))
		return
	}
	c.JSON(http.StatusOK, ProofPayloadResponse{
		PostContent: performer.GeneratePostPayload(),
		SignPayload: performer.GenerateSignPayload(),
		CreatedAt:   util.TimeToTimestampString(v.CreatedAt),
		Uuid:        v.Uuid.String(),
	})
}

func proofPayloadCheckRequest(req *ProofPayloadRequest) bool {
	return string(req.Action) != "" &&
		req.Platform != "" &&
		req.Identity != "" &&
		req.PublicKey != ""

}
