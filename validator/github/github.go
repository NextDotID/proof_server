package github

import (
	"github.com/nextdotid/proof-server/types"
	"github.com/nextdotid/proof-server/validator"
)

type Github validator.Base

func Init() {
	if validator.PlatformFactories == nil {
		validator.PlatformFactories = make(map[types.Platform]func(validator.Base) validator.IValidator)
	}
	validator.PlatformFactories[types.Platforms.Github] = func(base validator.Base) validator.IValidator {
		return Github(base)
	}
}

func (ghub Github) GeneratePostPayload() (post string) {
	// TODO: implement me
	return ""
}

func (ghub Github) GenerateSignPayload() (payload string) {
	// TODO: implement me
	return ""
}

func (ghub Github) Validate() (err error) {
	// TODO: implement me
	return nil
}
