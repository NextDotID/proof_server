package controller

import (
	"testing"

	"github.com/nextdotid/proof-server/model"
	"github.com/nextdotid/proof-server/types"
	"github.com/nextdotid/proof-server/util/crypto"
	"github.com/nextdotid/proof-server/validator"
	"github.com/stretchr/testify/assert"
)

func Test_proofQuery(t *testing.T) {
	t.Run("smoke", func(t *testing.T) {
		before_each(t)
		resp := ProofQueryResponse{}

		APITestCall(Engine, "GET", "/v1/proof?platform=twitter&identity=yeiwb", "", &resp)
		assert.Equal(t, 0, len(resp.IDs))
	})

	t.Run("success", func(t *testing.T) {
		before_each(t)

		pubkey, _ := crypto.StringToPubkey("0x028c3cda474361179d653c41a62f6bbb07265d535121e19aedf660da2924d0b1e3")
		_, err := model.ProofCreateFromValidator(&validator.Base{
			Platform:      types.Platforms.Twitter,
			Previous:      "",
			Action:        types.Actions.Create,
			Pubkey:        pubkey,
			Identity:      "yeiwb",
			ProofLocation: "1469221200140574721",
		})
		assert.Nil(t, err)

		resp := ProofQueryResponse{}
		APITestCall(Engine, "GET", "/v1/proof?platform=twitter&identity=yeiwb", "", &resp)
		assert.Equal(t, 2, len(resp.IDs))

		empty_resp := ProofQueryResponse{}
		APITestCall(Engine, "GET", "/v1/proof?platform=keybase&identity=yeiwb", "", &empty_resp)
		assert.Equal(t, 0, len(empty_resp.IDs))
	})
}
