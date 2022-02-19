package controller

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nextdotid/proof-server/model"
	"github.com/nextdotid/proof-server/types"
	"github.com/nextdotid/proof-server/util/crypto"
	"github.com/nextdotid/proof-server/validator"
	"golang.org/x/xerrors"
)

type KVPayloadRequest struct {
	Persona string        `json:"persona"`
	Changes model.KVPatch `json:"changes"`
}

type KVPayloadResponse struct {
	SignPayload string `json:"sign_payload"`
}

func kvPatchPayload(c *gin.Context) {
	req := KVPayloadRequest{}
	err := c.BindJSON(&req)
	if err != nil {
		errorResp(c, http.StatusBadRequest, xerrors.Errorf("body parse error"))
		return
	}
	changes_json, _ := json.Marshal(req.Changes)
	pubkey, err := crypto.StringToPubkey(req.Persona)
	if err != nil {
		errorResp(c, http.StatusBadRequest, xerrors.Errorf("error when marshaling persona"))
		return
	}
	pc, err := model.ProofChainFindLatest(crypto.CompressedPubkeyHex(pubkey))
	if err != nil {
		errorResp(c, http.StatusInternalServerError, xerrors.Errorf("error when fetching proof chain"))
		return
	}

	base := validator.Base{
		Platform: types.Platforms.KV,
		Action: types.Actions.KV,
		Pubkey: pubkey,
		Text:   string(changes_json),
	}
	if pc != nil {
		base.Previous = pc.Signature
	}

	performer := validator.BaseToInterface(&base)
	if performer == nil {
		errorResp(c, http.StatusBadRequest, xerrors.New("unknown platform"))
		return
	}

	c.JSON(http.StatusOK, KVPayloadResponse{
		SignPayload: performer.GenerateSignPayload(),
	})
}
