package model

import (
	"encoding/json"

	"golang.org/x/xerrors"
	"gorm.io/datatypes"
)

// KV is a model of persona-related key-value pair.
type KV struct {
	ID           int64             `gorm:"primarykey"`
	Persona      string            `gorm:"index; not null"`
	Content      datatypes.JSONMap `gorm:"default:'{}'"`
	ProofChainID int64             `gorm:"column:proof_chain_id; index"`
	ProofChain   *ProofChain
}

type KVContent map[string]interface{}
type KVPatch struct {
	Set KVContent `json:"set"`
	Del []string  `json:"del"`
}

func (KV) TableName() string {
	return "kv"
}

func (kv *KV) GetContent() (KVContent, error) {
	result := KVContent{}
	json_bytes, err := kv.Content.MarshalJSON()
	if err != nil {
		return result, xerrors.Errorf("%w", err)
	}

	err = json.Unmarshal(json_bytes, &result)
	if err != nil {
		return KVContent{}, xerrors.Errorf("%w", err)
	}

	return result, nil
}

func (kv *KV) ToJSONString() string {
	result, err := kv.Content.MarshalJSON()
	if err != nil {
		return ""
	}
	return string(result)
}

func KVApplyPatchFromProofChain(pc *ProofChain) (error) {
	kv, err := KVFindByPersona(pc.Persona)
	if err != nil {
		return xerrors.Errorf("%w", err)
	}
	if kv == nil {
		kv = &KV{
			Persona:      pc.Persona,
			ProofChainID: pc.ID,
			ProofChain:   pc,
		}
		DB.Create(&kv)
	}

	return kv.ApplyPatch(pc.UnmarshalExtra().KVPatch)
}

func (kv *KV) ApplyPatch(patch KVPatch) (error) {
	content, err := kv.GetContent()
	if err != nil {
		return xerrors.Errorf("%w", err)
	}

	// Set
	for k, v := range patch.Set {
		content[k] = v
	}

	// Del
	for _, del_key := range patch.Del {
		delete(content, del_key)
	}

	return kv.OverrideContent(content)
}

func (kv *KV) OverrideContent(new_content KVContent) (error) {
	kv.Content = datatypes.JSONMap(new_content)
	tx := DB.Save(kv)
	if tx.Error != nil {
		return xerrors.Errorf("error when updating KV: %w", tx.Error)
	}

	return nil
}

// KVFindByPersona accepts `*ecdsa.Pubkey` and `string` types.
func KVFindByPersona(persona interface{}) (*KV, error) {
	pubkey := MarshalPersona(persona)
	if pubkey == "" {
		return nil, xerrors.Errorf("pubkey not recognized")
	}

	result := KV{
		Persona: pubkey,
	}

	tx := DB.Find(&result)
	if tx.Error != nil {
		return nil, xerrors.Errorf("%w", tx.Error)
	}
	if tx.RowsAffected == int64(0) {
		return nil, nil
	}

	return &result, nil
}
