package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nextdotid/proof-server/model"
	"github.com/nextdotid/proof-server/types"
	"github.com/nextdotid/proof-server/util/crypto"
	"github.com/nextdotid/proof-server/validator"
	"golang.org/x/xerrors"
)

type ProofPayloadRequest struct {
	Action    types.Action   `json:"action"`
	Platform  types.Platform `json:"platform"`
	Identity  string         `json:"identity"`
	PublicKey string         `json:"public_key"`
	Extra     ProofPayloadRequestExtra `json:"extra"`
}

type ProofPayloadResponse struct {
	PostContent string `json:"post_content"`
	SignPayload string `jsoN:"sign_payload"`
}

type ProofPayloadRequestExtra struct {
	EthereumWalletSignature string `json:"wallet_signature"`
}

func proofPayload(c *gin.Context) {
	req := &ProofPayloadRequest{}
	err := c.BindJSON(req)
	if err != nil {
		errorResp(c, http.StatusBadRequest, err)
		return
	}
	if !proofPayloadCheckRequest(req) {
		errorResp(c, http.StatusBadRequest, xerrors.New("param invalid"))
		return
	}

	parsed_pubkey, err := crypto.StringToPubkey(req.PublicKey)
	if err != nil {
		errorResp(c, http.StatusBadRequest, xerrors.New("public key not recognized"))
		return
	}

	previous_proof, err := model.ProofFindLatest(crypto.CompressedPubkeyHex(parsed_pubkey))
	if err != nil {
		errorResp(c, http.StatusInternalServerError, xerrors.New("database error"))
		return
	}

	var previous_signature string
	if previous_proof == nil {
		previous_signature = ""
	} else {
		previous_signature = previous_proof.Signature
	}

	v := validator.Base{
		Platform: req.Platform,
		Previous: previous_signature,
		Action:   req.Action,
		Pubkey:   parsed_pubkey,
		Identity: req.Identity,
		Extra: map[string]string{
			"wallet_signature": req.Extra.EthereumWalletSignature,
		},
	}
	performer_factory, ok := validator.Platforms[req.Platform]
	if !ok {
		errorResp(c, http.StatusBadRequest, xerrors.New("unknown platform"))
		return
	}
	performer := performer_factory(v)
	c.JSON(http.StatusOK, ProofPayloadResponse{
		PostContent: performer.GeneratePostPayload(),
		SignPayload:performer.GenerateSignPayload(),
	})
}

func proofPayloadCheckRequest(req *ProofPayloadRequest) bool {
	return string(req.Action) != "" &&
		req.Platform != "" &&
		req.Identity != "" &&
		req.PublicKey != ""

}