package ens

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/google/uuid"
	"github.com/nextdotid/proof_server/config"
	"github.com/nextdotid/proof_server/types"
	"github.com/nextdotid/proof_server/util"
	"github.com/nextdotid/proof_server/util/crypto"
	"github.com/nextdotid/proof_server/validator"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	ensv3 "github.com/wealdtech/go-ens/v3"
	"golang.org/x/xerrors"
)

const ensKey = "id.next.proof"

var client *ethclient.Client

type TXTPayload struct {
	Version   uint
	Signature string
	CreatedAt time.Time
	uuid      uuid.UUID
	Previous  *string
}

type ENS struct {
	*validator.Base
}

const (
	TXT_PAYLOAD_V1 = "ps:true;v:1;sig:%s;ca:%d;uuid:%s;prev:%s"
)

var (
	l = logrus.WithFields(logrus.Fields{"module": "validator", "validator": "ens"})
)

func Init() {
	initClient()
	if validator.PlatformFactories == nil {
		validator.PlatformFactories = make(map[types.Platform]func(*validator.Base) validator.IValidator)
	}
	validator.PlatformFactories[types.Platforms.ENS] = func(base *validator.Base) validator.IValidator {
		ens := ENS{base}
		return &ens
	}

}

func (ens *ENS) GeneratePostPayload() (post map[string]string) {
	var previous string
	if ens.Previous != "" {
		previous = ens.Previous
	} else {
		previous = "null"
	}
	return map[string]string{
		"default": fmt.Sprintf(TXT_PAYLOAD_V1, "%SIG_BASE64%", ens.CreatedAt.Unix(), ens.Uuid.String(), previous),
	}
}

func (ens *ENS) GenerateSignPayload() (payload string) {
	payloadStruct := validator.H{
		"action":     string(ens.Action),
		"identity":   strings.ToLower(ens.Identity),
		"platform":   string(types.Platforms.ENS),
		"prev":       nil,
		"created_at": util.TimeToTimestampString(ens.CreatedAt),
		"uuid":       ens.Uuid.String(),
	}
	if ens.Previous != "" {
		payloadStruct["prev"] = ens.Previous
	}
	payloadBytes, err := json.Marshal(payloadStruct)
	if err != nil {
		l.Warnf("Error when marshaling struct: %s", err.Error())
		return ""
	}

	return string(payloadBytes)
}

func (ens *ENS) Validate() (err error) {
	initClient()
	// domain name is case-insensitive
	ens.Identity = strings.ToLower(ens.Identity)
	ens.AltID = ens.Identity
	ens.SignaturePayload = ens.GenerateSignPayload()
	resolver, err := ensv3.NewDNSResolver(client, ens.Identity)
	if err != nil {
		return xerrors.Errorf("While finding the ens resolver for the given name: %v", err)
	}
	nh, err := ensv3.NameHash(ens.Identity)
	if err != nil {
		return xerrors.Errorf("While hashing the ens name: %v", err)
	}
	txt, err := resolver.Contract.Text(&bind.CallOpts{}, nh, ensKey)
	if err != nil {
		return xerrors.Errorf("matched TXT record couldn't be retrieved: %v", err)
	}
	txtData := string(txt)
	payload, _ := parseTxt(txtData)
	ens.Text = txtData
	ens.Signature, err = base64.StdEncoding.DecodeString(payload.Signature)
	if err != nil {
		return xerrors.New("sig in TXT record cannot be recognized.")
	}

	return crypto.ValidatePersonalSignature(ens.SignaturePayload, ens.Signature, ens.Pubkey)
}

func (ens *ENS) GetAltID() string {
	return ens.AltID
}

func parseTxt(txtField string) (result TXTPayload, err error) {
	kv := make(map[string]string)

	// txtField = "\"ps:true;v:1;sig:3QgQUPrPiBloBev8uf1wyjpa4roK4xjN2OXeBpqQFYMOHFo+blMR0Ppyc/JVj0jtdLDGBTrOFdOJPMfvUXZkwAE=;ca:1664267795;uuid:80c98711-f4f6-43c7-b05c-8d86372f6131;prev:null\""
	lo.ForEach(strings.Split(strings.Trim(txtField, "\""), ";"), func(combined string, i int) {
		pair := strings.Split(combined, ":")
		if len(pair) != 2 {
			err = xerrors.Errorf("TXT payload format error in %s", combined)
			return
		}
		kv[pair[0]] = pair[1]
	})
	if err != nil {
		return TXTPayload{}, err
	}
	if !lo.Every(lo.Keys(kv), []string{"ps", "v", "sig", "ca", "uuid", "prev"}) {
		return TXTPayload{}, xerrors.New("TXT payload not recognized: field missing")
	}
	if kv["ps"] != "true" {
		return TXTPayload{}, xerrors.New("TXT payload not recognized")
	}
	if kv["v"] != "1" {
		return TXTPayload{}, xerrors.New("TXT payload version not recognized")
	}

	uuid, err := uuid.Parse(kv["uuid"])
	if err != nil {
		return TXTPayload{}, err
	}
	createdAt, err := util.TimestampStringToTime(kv["ca"])
	if err != nil {
		return TXTPayload{}, err
	}

	var previous *string = nil
	if kv["prev"] != "null" {
		prev := kv["prev"]
		previous = &prev
	}

	return TXTPayload{
		Version:   1,
		Signature: kv["sig"],
		CreatedAt: createdAt,
		uuid:      uuid,
		Previous:  previous,
	}, nil
}

func initClient() error {
	if client != nil {
		return nil
	}
	var err error
	client, err = ethclient.Dial(config.C.Platform.Ethereum.RPCServer)
	return err
}
