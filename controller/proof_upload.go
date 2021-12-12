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
	"github.com/sirupsen/logrus"
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

	if err := validateProof(req, proof, pubkey); err != nil {
		errorResp(c, 400 , xerrors.Errorf("%w", err))
		return
	}

	if err := applyUpload(req, proof, pubkey); err != nil {
		errorResp(c, 400 , xerrors.Errorf("%w", err))
		return
	}

	c.JSON(http.StatusCreated, gin.H{})
}


func validateProof(req ProofUploadRequest, prev *model.Proof, pubkey *ecdsa.PublicKey) error {
	proof_signature := ""
	if prev != nil {
		proof_signature = prev.Signature
	}

	switch req.Platform {
	case types.Platforms.Twitter:
		tweet := twitter.Twitter{
			Base: validator.Base{
				Previous:      proof_signature,
				Action:        req.Action,
				Pubkey:        pubkey,
				Identity:      req.Identity,
				ProofLocation: req.ProofLocation,
			},
		}
		return tweet.Validate()
	case types.Platforms.Keybase:
		kb := keybase.Keybase{
			Base: validator.Base{
				Previous:      proof_signature,
				Action:        req.Action,
				Pubkey:        pubkey,
				Identity:      req.Identity,
			},
		}
		return kb.Validate()
	default:
		return xerrors.Errorf("platform not supported: %s", string(req.Platform))
	}
}

func applyUpload(req ProofUploadRequest, prev *model.Proof, pubkey *ecdsa.PublicKey) error {
	switch req.Action {
	case types.Actions.Create:
		return generateProof(req, prev, pubkey)
	case types.Actions.Delete:
		return deleteProof(req, prev, pubkey)
	default:
		return xerrors.Errorf("Unknown action: %s", string(req.Action))
	}
}

func generateProof(req ProofUploadRequest, prev *model.Proof, pubkey *ecdsa.PublicKey) error {
	prev_id := uint(0)
	if prev != nil {
		prev_id = prev.ID
	}
	// FIXME: Proof creation should be an instance method
	proof := model.Proof{
		PreviousProof: prev_id,
		Persona:       "0x" + crypto.CompressedPubkeyHex(pubkey),
		Platform:      req.Platform,
		Identity:      req.Identity,
		Location:      req.ProofLocation,
		// FIXME: Signature handling logic
		Signature:     "test123",
	}
	logrus.Warnf("%+v", proof)
	tx := model.DB.Create(&proof)
	logrus.Warnf("TX: %+v", tx)
	if tx.Error != nil {
		return xerrors.Errorf("%w", tx.Error)
	}
	return nil
}

func deleteProof(req ProofUploadRequest, prev *model.Proof, pubkey *ecdsa.PublicKey) error {
	// FIXME: impelement this
	return nil
}
