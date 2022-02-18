package model

import "time"

// KV is a model of persona-related key-value pair.
type KV struct {
	ID           int64  `gorm:"primarykey"`
	Persona      string `gorm:"index; not null"`
	Key          string `gorm:"index"`
	Value        string
	ProofChainID int64 `gorm:"column:proof_chain_id; index"`
	ProofChain   *ProofChain
}

func (KV) TableName() string {
	return "kv"
}
