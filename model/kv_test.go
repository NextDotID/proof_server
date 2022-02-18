package model

import (
	"database/sql"
	"testing"

	"github.com/nextdotid/proof-server/types"
	"github.com/stretchr/testify/assert"
	"gorm.io/datatypes"
)

var (
	kv_test_data = KV{
		Persona:      "028c3cda474361179d653c41a62f6bbb07265d535121e19aedf660da2924d0b1e3",
		Content:      datatypes.JSONMap{
			"this": "is",
			"a": []string{"test", "data"},
		},
	}
)

func test_create_kv_and_proof_chain() (*KV, *ProofChain) {
	kv := kv_test_data
	pc := ProofChain{
		Action:           types.Actions.KV,
		Persona:          kv.Persona,
		Identity:         "",
		Platform:         "",
		Extra:            datatypes.JSONMap{
			"kv_patch": map[string]interface{}{
				"set": map[string]interface{}{
					"a": []string{"test", "data"},
					"this": "is",
				},
				"del": []string{"undefined.key"},
			},
		},
		PreviousID:       sql.NullInt64{},
		Previous:         &ProofChain{},
	}
	DB.Create(&pc)
	kv.ProofChain = &pc
	DB.Create(&kv)
	return &kv, &pc
}

func Test_ToJSONSring(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)

		content_json := kv_test_data.ToJSONString()
		assert.Equal(t, "{\"a\":[\"test\",\"data\"],\"this\":\"is\"}", content_json)
	})
}

func Test_GetContent(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)

		content_struct, err := kv_test_data.GetContent()
		assert.Nil(t, err)
		assert.Equal(t, content_struct["this"], "is")
	})
}

func Test_ApplyPatch(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)
		kv := kv_test_data
		tx := DB.Create(&kv)
		assert.Nil(t, tx.Error)

		patch := KVPatch{
			Set: map[string]interface{}{
				"test": 123,
			},
			Del: []string{"a"},
		}
		err := kv.ApplyPatch(patch)
		assert.Nil(t, err)

		content, _ := kv.GetContent()
		assert.Equal(t, float64(123), content["test"])
	})
}
