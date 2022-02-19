package controller

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nextdotid/proof-server/model"
	"github.com/nextdotid/proof-server/types"
	"github.com/nextdotid/proof-server/util/crypto"
	"github.com/nextdotid/proof-server/validator"
	"golang.org/x/xerrors"
)

type KVPatchRequest struct {
	Persona         string        `json:"persona"`
	SignatureBase64 string        `json:"signature"`
	Changes         model.KVPatch `jsonL:"changes"`
}

func kvPatch(c *gin.Context) {
	req := KVPatchRequest{}
	err := c.BindJSON(&req)
	if err != nil {
		errorResp(c, http.StatusBadRequest, xerrors.Errorf("parsing body error"))
		return
	}
	pubkey, err := crypto.StringToPubkey(req.Persona)
	if err != nil {
		errorResp(c, http.StatusBadRequest, xerrors.Errorf("decoding persona error"))
		return
	}

	sig, err := base64.StdEncoding.DecodeString(req.SignatureBase64)
	if err != nil {
		errorResp(c, http.StatusBadRequest, xerrors.Errorf("decoding signature error"))
		return
	}

	changes_json, _ := json.Marshal(req.Changes)
	base := validator.Base{
		Platform:  types.Platforms.KV,
		Previous:  "",
		Action:    types.Actions.KV,
		Pubkey:    pubkey,
		Signature: sig,
		Text:      string(changes_json),
		Extra: map[string]interface{}{
			"kv_patch": string(changes_json),
		},
	}

	previous_pc, err := model.ProofChainFindLatest(crypto.CompressedPubkeyHex(pubkey))
	if err != nil {
		errorResp(c, http.StatusBadRequest, xerrors.Errorf("finding previous proof error"))
		return
	}
	if previous_pc != nil {
		base.Previous = previous_pc.Signature
	}

	performer_factory, ok := validator.PlatformFactories[types.Platforms.KV]
	if !ok {
		errorResp(c, http.StatusBadRequest, xerrors.Errorf("KV function not supported by this server."))
		return
	}

	performer := performer_factory(&base)
	if err := performer.Validate(); err != nil {
		errorResp(c, http.StatusBadRequest, xerrors.Errorf("validation error: %w", err))
		return
	}

	if err := applyUpload(&base); err != nil {
		errorResp(c, http.StatusInternalServerError, xerrors.Errorf("%w", err))
		return
	}

	c.JSON(http.StatusCreated, gin.H{})
}
