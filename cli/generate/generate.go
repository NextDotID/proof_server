package generate

import (
	"crypto/ecdsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	"github.com/nextdotid/proof-server/controller"
	"github.com/nextdotid/proof-server/types"
	"github.com/nextdotid/proof-server/util/crypto"
	"github.com/nextdotid/proof-server/validator"
	"github.com/nextdotid/proof-server/validator/ethereum"
	"github.com/nextdotid/proof-server/validator/github"
	"github.com/nextdotid/proof-server/validator/keybase"
	"github.com/nextdotid/proof-server/validator/twitter"
	"regexp"
	"strings"
	"time"
)

var post_regex = regexp.MustCompile("%SIG_BASE64%")

func init_validators() {
	twitter.Init()
	github.Init()
	keybase.Init()
	ethereum.Init()
}

func get_action(action string) types.Action {
	switch action {
	case "create":
		return types.Actions.Create
	case "delete":
		return types.Actions.Delete
	default:
		panic(fmt.Sprintf("Unknown action: %s", action))
	}
}

func GeneratePayload(sk string, platform string, identity string, prev string, action string, eth_sk string) {
	init_validators()
	secret_key, err := ethcrypto.HexToECDSA(sk)
	if err != nil {
		panic(err)
	}

	pl := types.Platform(platform)
	platform_factory, ok := validator.PlatformFactories[pl]
	if !ok {
		panic(fmt.Sprintf("Platform %s not found", pl))
	}

	base := validator.Base{
		Platform:  pl,
		Previous:  prev,
		Action:    get_action(action),
		Pubkey:    &secret_key.PublicKey,
		Identity:  strings.ToLower(identity),
		CreatedAt: time.Now(),
		Uuid:      uuid.New(),
	}
	validator := platform_factory(&base)

	sign_payload := validator.GenerateSignPayload()
	fmt.Printf("Sign payload: vvvvvvvvv\n%s\n^^^^^^^^^^^^^^^\n\n", sign_payload)

	if pl == types.Platforms.Ethereum {
		generate_ethereum(sign_payload, secret_key, action, eth_sk)
		return
	}
	raw_post := validator.GeneratePostPayload()
	signature, err := crypto.SignPersonal([]byte(sign_payload), secret_key)
	if err != nil {
		panic(err)
	}
	for lang_code, payload := range raw_post {
		fmt.Printf(
			"Post payload [%s]: vvvvvvv\n%s\n^^^^^^^^^^\n\n",
			lang_code,
			string(post_regex.ReplaceAll([]byte(payload), []byte(base64.StdEncoding.EncodeToString(signature)))),
		)
	}
}

func generate_ethereum(sign_payload string, persona_sk *ecdsa.PrivateKey, action string, eth_sk string) {
	persona_sig, _ := crypto.SignPersonal([]byte(sign_payload), persona_sk)
	fmt.Printf("Persona sig: vvvvvvvvvv\n%s\n^^^^^^^^^^^^^^^\n\n", base64.StdEncoding.EncodeToString(persona_sig))

	wallet_sk, err := ethcrypto.HexToECDSA(eth_sk)
	if err != nil {
		panic(fmt.Sprintf("ETH secret key failed: %s", err.Error()))
	}
	wallet_sig, _ := crypto.SignPersonal([]byte(sign_payload), wallet_sk)
	fmt.Printf("Wallet sig: vvvvvvvvvv\n%s\n^^^^^^^^^^^^^^^\n\n", base64.StdEncoding.EncodeToString(wallet_sig))

	persona_pk := &persona_sk.PublicKey

	req := controller.ProofUploadRequest{
		Action:    types.Action(action),
		Platform:  types.Platforms.Ethereum,
		Identity:  ethcrypto.PubkeyToAddress(wallet_sk.PublicKey).String(),
		PublicKey: "0x" + crypto.CompressedPubkeyHex(persona_pk),
		Extra: controller.ProofUploadRequestExtra{
			EthereumWalletSignature: base64.StdEncoding.EncodeToString(wallet_sig),
			Signature:               base64.StdEncoding.EncodeToString(persona_sig),
		},
	}

	req_json, _ := json.Marshal(req)
	fmt.Printf("POST /v1/proof/payload request:\n\n%s\n\n", req_json)
}
