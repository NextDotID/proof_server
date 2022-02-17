package kv

import (
	"github.com/nextdotid/proof-server/types"
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
		kv := KV { base }
		return &kv
	}
}

func (kv *KV) GeneratePostPayload() (string) {
	return ""
}

func (kv *KV) GenerateSignPayload() (string) {
	return "" // TODO
}

func (kv *KV) Validate() (err error) {
	return nil // TODO
}
