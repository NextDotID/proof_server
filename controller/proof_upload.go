package controller

import (
	"crypto/ecdsa"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nextdotid/proof-server/model"
	"github.com/nextdotid/proof-server/types"
	"github.com/nextdotid/proof-server/util/crypto"
	"github.com/nextdotid/proof-server/validator"
	"github.com/nextdotid/proof-server/validator/keybase"
	"github.com/nextdotid/proof-server/validator/twitter"
	"golang.org/x/xerrors"
)

type ProofUploadRequest struct {
	Action        types.Action   `json:"action"`
	Platform      types.Platform `json:"platform"`
	Identity      string         `json:"identity"`
	ProofLocation string         `json:"proof_location"`
	PublicKey     string         `json:"public_key"`
}

func proofUpload(c *gin.Context) {
	req := ProofUploadRequest{}
	err := c.BindJSON(&req)
	if err != nil {
		errorResp(c, 400, xerrors.Errorf("parse request failed: %w", err))
		return
	}
	pubkey, err := crypto.StringToPubkey(req.PublicKey)
	if err != nil {
		errorResp(c, 400, xerrors.Errorf("%w", err))
		return
	}

	proof, err := model.ProofFindLatest(crypto.CompressedPubkeyHex(pubkey))
	if err != nil {
		errorResp(c, 500, xerrors.Errorf("internal database error"))
		return
	}

	validator, err := validateProof(req, proof, pubkey)
	if err != nil {
		errorResp(c, 400, xerrors.Errorf("%w", err))
		return
	}

	if err := applyUpload(req, proof, &validator); err != nil {
		errorResp(c, 400, xerrors.Errorf("%w", err))
		return
	}

	c.JSON(http.StatusCreated, gin.H{})
}

func validateProof(req ProofUploadRequest, prev *model.Proof, pubkey *ecdsa.PublicKey) (validator.Base, error) {
	proof_signature := ""
	if prev != nil {
		proof_signature = prev.Signature
	}
	base := validator.Base{
		Platform:      req.Platform,
		Previous:      proof_signature,
		Action:        req.Action,
		Pubkey:        pubkey,
		Identity:      req.Identity,
		ProofLocation: req.ProofLocation,
	}

	switch req.Platform {
	case types.Platforms.Twitter:
		v_performer := twitter.Twitter(base)
		return base, v_performer.Validate()
	case types.Platforms.Keybase:
		v_performer := keybase.Keybase(base)
		return base, v_performer.Validate()
	default:
		return validator.Base{}, xerrors.Errorf("platform not supported: %s", string(req.Platform))
	}
}

func applyUpload(req ProofUploadRequest, prev *model.Proof, validator *validator.Base) error {
	switch req.Action {
	case types.Actions.Create:
		return generateProof(req, prev, validator)
	case types.Actions.Delete:
		return deleteProof(req, prev, validator)
	default:
		return xerrors.Errorf("Unknown action: %s", string(req.Action))
	}
}

func generateProof(req ProofUploadRequest, prev *model.Proof, validator *validator.Base) error {
	_, err := model.ProofCreateFromValidator(validator)
	return err
}

func deleteProof(req ProofUploadRequest, prev *model.Proof, validator *validator.Base) error {
	// FIXME: impelement this
	return nil
}
