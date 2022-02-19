package kv

import (
	"encoding/json"

	"github.com/nextdotid/proof-server/model"
	"github.com/nextdotid/proof-server/types"
	"github.com/nextdotid/proof-server/util/crypto"
	"github.com/nextdotid/proof-server/validator"
	"github.com/sirupsen/logrus"
)

type KV struct {
	*validator.Base
}

var (
	l = logrus.WithFields(logrus.Fields{"module": "validator", "validator": "kv"})
)

func Init() {
	if validator.PlatformFactories == nil {
		validator.PlatformFactories = make(map[types.Platform]func(*validator.Base) validator.IValidator)
	}
	validator.PlatformFactories[types.Platforms.KV] = func(base *validator.Base) validator.IValidator {
		kv := KV{base}
		return &kv
	}
}

func (kv *KV) GeneratePostPayload() string {
	return ""
}

func (kv *KV) GenerateSignPayload() string {
	patch := model.KVPatch{}
	err := json.Unmarshal([]byte(kv.Text), &patch)
	if err != nil {
		return ""
	}

	payload := validator.H{
		"action": types.Actions.KV,
		"set":    patch.Set,
		"del":    patch.Del,
		"prev":   nil,
	}
	if kv.Previous != "" {
		payload["prev"] = kv.Previous
	}

	result, _ := json.Marshal(payload)
	return string(result)
}

func (kv *KV) Validate() (err error) {
	sign_payload := kv.GenerateSignPayload()
	return crypto.ValidatePersonalSignature(sign_payload, kv.Signature, kv.Pubkey)
}
