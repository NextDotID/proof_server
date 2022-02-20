package model

import (
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/gin-gonic/gin"
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

func test_create_kv_and_proof_chain(t *testing.T) (*KV, *ProofChain) {
	kv := kv_test_data
	extra := gin.H{
		"kv_patch": map[string]interface{}{
			"set": map[string]interface{}{
				"a": []string{"test", "data"},
				"this": "is",
			},
			"del": []string{"undefined.key"},
		},
	}
	extra_bytes, _ := json.Marshal(extra)
	pc := ProofChain{
		Action:           types.Actions.KV,
		Persona:          kv.Persona,
		Identity:         "",
		Platform:         "",
		Extra:            datatypes.JSON(extra_bytes),
		PreviousID:       sql.NullInt64{},
		Previous:         &ProofChain{},
	}
	tx := DB.Create(&pc)
	assert.Nil(t, tx.Error)
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
		kv, _ := test_create_kv_and_proof_chain(t)

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
		assert.Equal(t, "is", content["this"])
		_, ok := content["a"]
		assert.False(t, ok)
	})
}

func Test_KVApplyPatchFromProofChain(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)
		orig_kv, pc := test_create_kv_and_proof_chain(t)

		new_kv, err := KVApplyPatchFromProofChain(pc)
		assert.Nil(t, err)
		assert.Equal(t, orig_kv.ID, new_kv.ID)
		content, _ := new_kv.GetContent()
		assert.Equal(t, "is", content["this"])
	})
}

func Test_KVFindByPersona(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)
		orig_kv, _ := test_create_kv_and_proof_chain(t)

		kv, err := KVFindByPersona(orig_kv.Persona)
		assert.Nil(t, err)
		assert.Equal(t, orig_kv.ID, kv.ID)
	})

	t.Run("not found", func(t *testing.T) {
		before_each(t)

		kv, err := KVFindByPersona(kv_test_data.Persona)
		assert.Nil(t, err)
		assert.Nil(t, kv)
	})
}
