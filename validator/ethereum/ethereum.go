package ethereum

import (
	"encoding/json"
	"regexp"

	"github.com/nextdotid/proof-server/validator"
	"github.com/sirupsen/logrus"
)

type Ethereum validator.Base

const (
	VALIDATE_TEMPLATE = `^{"eth_address":"0x([0-9a-fA-F]{40})","signature":"(.*)"}$`
)

var (
	l = logrus.WithFields(logrus.Fields{"module": "validator", "validator": "ethereum"})
	re = regexp.MustCompile(VALIDATE_TEMPLATE)
)

// Not used by etheruem (for now).
func (Ethereum) GeneratepostPayload() (post string) {
	return ""
}

func (et *Ethereum) GenerateSignPayload() (payload string) {
	payloadStruct := validator.H{
		"action": string(et.Action),
		"identity": et.Pubkey,
		"eth_address": et.Identity,
		"prev": nil,
	}
	if et.Previous != "" {
		payloadStruct["prev"] = et.Previous
	}
	payloadBytes, err := json.Marshal(payloadStruct)
	if err != nil {
		l.Warnf("Error when marshaling struct: %s", err.Error())
		return ""
	}

	return string(payloadBytes)
}
