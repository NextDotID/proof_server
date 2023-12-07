package tiktok

import (
	"github.com/nextdotid/proof_server/config"
	"github.com/nextdotid/proof_server/types"
	"github.com/nextdotid/proof_server/validator"
	"github.com/sirupsen/logrus"
)

type TikTok struct {
	*validator.Base
}

var (
	l = logrus.WithFields(logrus.Fields{"module": "validator", "validator": "tiktok"})
)

func Init() {
	if validator.PlatformFactories == nil {
		validator.PlatformFactories = make(map[types.Platform]func(*validator.Base) validator.IValidator)
	}
	cfg := config.C.Platform.TikTok
	if cfg.ClientKey == "" || cfg.ClientSecret == "" || cfg.AppID == "" {
		l.Warn("Config is missing. Skip initializing.")
		return
	}
	validator.PlatformFactories[types.Platforms.Solana] = func(base *validator.Base) validator.IValidator {
		tt := TikTok{base}
		return &tt
	}
}

func (tt *TikTok) GeneratePostPayload() (post map[string]string) {
	return map[string]string{} // TODO
}

func (tt *TikTok) GenerateSignPayload() (payload string) {
	return "" // TODO
}

func (tt *TikTok) Validate() (err error) {
	return nil // TODO
}

func (tt *TikTok) GetAltID() string{
	return "" // TODO
}
