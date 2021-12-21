package controller

import (
	"testing"

	"github.com/nextdotid/proof-server/model"
	"github.com/nextdotid/proof-server/types"
	"github.com/nextdotid/proof-server/util/crypto"
	"github.com/nextdotid/proof-server/validator"
	"github.com/stretchr/testify/assert"
)

const (
	persona = "0x028c3cda474361179d653c41a62f6bbb07265d535121e19aedf660da2924d0b1e3"
)

func insert_proof(t *testing.T) {
	persona := "0x028c3cda474361179d653c41a62f6bbb07265d535121e19aedf660da2924d0b1e3"
	pubkey, _ := crypto.StringToPubkey(persona)
	validators := []validator.Base{
		{
			Platform:      types.Platforms.Twitter,
			Previous:      "",
			Action:        types.Actions.Create,
			Pubkey:        pubkey,
			Identity:      "yeiwb",
			ProofLocation: "1469221200140574721",
			Signature:     []byte{1},
		},
		{
			Platform:      types.Platforms.Ethereum,
			Previous:      "0x01",
			Action:        types.Actions.Create,
			Pubkey:        pubkey,
			Identity:      "0xd5F630652D4a8a5f95cda3738CE9f43fa26E764F",
			ProofLocation: "",
			Signature:     []byte{2},
			Extra:         map[string]string{
				"ethereum_pubkey": "0x04ae5933a45605e7fff23cd010455911c1f0194479438859af5140d749937e53fd935d768efa9229ae8be3314631e945c56f915778ad4565b4efafcd13864e2fd7",
			},
		},
	}

	for _, b := range validators {
		_, err := model.ProofCreateFromValidator(&b)
		assert.Nil(t, err)
	}
}

func Test_proofQuery(t *testing.T) {
	t.Run("smoke", func(t *testing.T) {
		before_each(t)
		resp := ProofQueryResponse{}

		APITestCall(Engine, "GET", "/v1/proof?platform=twitter&identity=yeiwb", "", &resp)
		assert.Equal(t, 0, len(resp.IDs))
	})

	t.Run("success", func(t *testing.T) {
		before_each(t)
		insert_proof(t)

		resp := ProofQueryResponse{}
		APITestCall(Engine, "GET", "/v1/proof?platform=twitter&identity=yeiwb", "", &resp)
		assert.Equal(t, 1, len(resp.IDs))
		found := resp.IDs[0]
		assert.Equal(t, persona, found.Persona)
		assert.Equal(t, 1, len(found.Proofs))

		partial_resp := ProofQueryResponse{}
		APITestCall(Engine, "GET", "/v1/proof?platform=twitter&identity=eiw", "", &partial_resp)
		assert.Equal(t, 1, len(resp.IDs))
		found = partial_resp.IDs[0]
		assert.Equal(t, persona, found.Persona)
		assert.Equal(t, 1, len(found.Proofs))

		empty_resp := ProofQueryResponse{}
		APITestCall(Engine, "GET", "/v1/proof?platform=keybase&identity=yeiwb", "", &empty_resp)
		assert.Equal(t, 0, len(empty_resp.IDs))
	})
}

func Test_performProofQuery(t *testing.T) {

}
