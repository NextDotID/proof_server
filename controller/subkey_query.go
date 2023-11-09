package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/nextdotid/proof_server/model"
	"github.com/nextdotid/proof_server/types"
	"github.com/nextdotid/proof_server/util/crypto"
	"github.com/samber/lo"
	"golang.org/x/xerrors"
)

type subkeyQueryRequest struct {
	Avatar    string `form:"avatar"`
	PublicKey string `form:"public_key"`
	Algorithm string `form:"algorithm"`
}

type subkeyQueryResponse struct {
	Subkeys []subkeyQueryResponseSingle `json:"subkeys"`
}

type subkeyQueryResponseSingle struct {
	Avatar    string                `json:"avatar"`
	Algorithm types.SubkeyAlgorithm `json:"algorithm"`
	PublicKey string                `json:"public_key"`
	Name      string                `json:"name"`
	RP_ID     string                `json:"rp_id"`
	CreatedAt int64                 `json:"created_at"`
}

// GET /v1/subkey
func subkeyQuery(c *gin.Context) {
	req := subkeyQueryRequest{}
	var err error
	if err = c.BindQuery(&req); err != nil {
		errorResp(c, 400, err)
		return
	}
	if err = subkeyQueryRequestValid(&req); err != nil {
		errorResp(c, 400, err)
		return
	}

	var subkeys []model.Subkey
	if req.Avatar != "" {
		subkeys, err = subkeyQueryAvatar(&req)
	} else {
		subkeys, err = subkeyQuerySubkey(&req)
	}
	if err != nil {
		errorResp(c, 500, err)
		return
	}
	response := subkeyQueryResponse{
		Subkeys: lo.Map(subkeys, func(s model.Subkey, index int) subkeyQueryResponseSingle {
			return subkeyQueryResponseSingle{
				Avatar:    s.Avatar,
				Algorithm: s.Algorithm,
				PublicKey: s.PublicKey,
				Name:      s.Name,
				RP_ID:     s.RP_ID,
				CreatedAt: s.CreatedAt.Unix(),
			}
		}),
	}
	c.JSON(200, response)
}

func subkeyQueryAvatar(req *subkeyQueryRequest) (subkeys []model.Subkey, err error) {
	subkeys = make([]model.Subkey, 0)
	tx := model.ReadOnlyDB.Model(&model.Subkey{})
	result := tx.Where("avatar", req.Avatar).Find(&subkeys)
	if result.Error != nil {
		return []model.Subkey{}, result.Error
	}
	return subkeys, nil
}

func subkeyQuerySubkey(req *subkeyQueryRequest) (subkeys []model.Subkey, err error) {
	subkeys = make([]model.Subkey, 0)
	tx := model.ReadOnlyDB.Model(&model.Subkey{})
	result := tx.Where("algorithm", req.Algorithm).Where("public_key", req.PublicKey).Find(&subkeys)
	if result.Error != nil {
		return []model.Subkey{}, result.Error
	}
	return subkeys, nil
}

func subkeyQueryRequestValid(req *subkeyQueryRequest) error {
	if req.Avatar == "" && req.PublicKey == "" {
		return xerrors.New("Avatar or public_key should be given")
	}
	if req.Avatar != "" {
		_, err := crypto.StringToSecp256k1Pubkey(req.Avatar)
		if err != nil {
			return xerrors.Errorf("Error when parsing avatar: %w", err)
		}
	} else {
		switch req.Algorithm {
		case string(types.SubkeyAlgorithms.Secp256K1):
			_, err := crypto.StringToSecp256k1Pubkey(req.PublicKey)
			if err != nil {
				return xerrors.Errorf("Error when parsing subkey: %w", err)
			}
		case string(types.SubkeyAlgorithms.Secp256R1):
			_, err := crypto.StringToSecp256r1Pubkey(req.PublicKey)
			if err != nil {
				return xerrors.Errorf("Error when parsing subkey: %w", err)
			}
		default:
			return xerrors.New("One of avatar or algorighm should be given.")
		}
	}

	return nil
}
