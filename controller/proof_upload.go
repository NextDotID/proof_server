package controller

import (
	"crypto/ecdsa"
	"encoding/base64"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nextdotid/proof-server/model"
	"github.com/nextdotid/proof-server/types"
	"github.com/nextdotid/proof-server/util"
	mycrypto "github.com/nextdotid/proof-server/util/crypto"
	"github.com/nextdotid/proof-server/validator"
	"golang.org/x/xerrors"
)

type ProofUploadRequest struct {
	Action        types.Action            `json:"action"`
	Platform      types.Platform          `json:"platform"`
	Identity      string                  `json:"identity"`
	ProofLocation string                  `json:"proof_location"`
	PublicKey     string                  `json:"public_key"`
	Uuid          string                  `json:"uuid"`
	CreatedAt     string                  `json:"created_at"`
	Extra         ProofUploadRequestExtra `json:"extra"`
}

type ProofUploadRequestExtra struct {
	Signature               string `json:"signature"`
	EthereumWalletSignature string `json:"wallet_signature"`
}

func proofUpload(c *gin.Context) {
	req := ProofUploadRequest{}
	err := c.BindJSON(&req)
	if err != nil {
		errorResp(c, 400, xerrors.Errorf("parse request failed: %w", err))
		return
	}
	pubkey, err := mycrypto.StringToPubkey(req.PublicKey)
	if err != nil {
		errorResp(c, 400, xerrors.Errorf("%w", err))
		return
	}

	previous_pc, err := model.ProofChainFindLatest(mycrypto.CompressedPubkeyHex(pubkey))
	if err != nil {
		errorResp(c, 500, xerrors.Errorf("internal database error"))
		return
	}

	validator, err := validateProof(req, previous_pc, pubkey)
	if err != nil {
		errorResp(c, 400, xerrors.Errorf("%w", err))
		return
	}

	if err = applyUpload(&validator); err != nil {
		errorResp(c, 400, xerrors.Errorf("%w", err))
		return
	}

	c.JSON(http.StatusCreated, gin.H{})
}

func validateProof(req ProofUploadRequest, prev *model.ProofChain, pubkey *ecdsa.PublicKey) (validator.Base, error) {
	prev_signature := ""
	if prev != nil {
		prev_signature = prev.Signature
	}

	performer_factory, ok := validator.PlatformFactories[req.Platform]
	if !ok {
		return validator.Base{}, xerrors.Errorf("platform not supported: %s", string(req.Platform))
	}
	created_at, err := util.TimestampStringToTime(req.CreatedAt)
	if err != nil {
		return validator.Base{}, xerrors.Errorf("error when parsing created_at: %s not recognized", req.CreatedAt)
	}
	parsed_uuid, err := uuid.Parse(req.Uuid)
	if err != nil {
		return validator.Base{}, xerrors.Errorf("error when parsing uuid: %s not recognized", req.Uuid)
	}
	base := validator.Base{
		Platform:      req.Platform,
		Previous:      prev_signature,
		Action:        req.Action,
		Pubkey:        pubkey,
		Identity:      req.Identity,
		ProofLocation: req.ProofLocation,
		CreatedAt:     created_at,
		Uuid:          parsed_uuid,
	}

	if req.Extra.Signature != "" || req.Platform == types.Platforms.Ethereum {
		extra := map[string]string{}
		extra["wallet_signature"] = req.Extra.EthereumWalletSignature
		base.Extra = extra

		persona_sig, err := base64.StdEncoding.DecodeString(req.Extra.Signature)
		if err != nil {
			return validator.Base{}, xerrors.Errorf("error when decoding persona signature: %w", err)
		}
		base.Signature = persona_sig
	}

	performer := performer_factory(&base)
	return base, performer.Validate()
}

func applyUpload(validator *validator.Base) error {
	pc, err := model.ProofChainCreateFromValidator(validator)
	if err != nil {
		return xerrors.Errorf("%w", err)
	}

	err = pc.Apply()
	if err != nil {
		return xerrors.Errorf("%w", err)
	}

	return nil
}
