package main

import (
	"crypto/ecdsa"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	"github.com/nextdotid/proof-server/cmd/cli/query"
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

var (
	flag_platform       = flag.String("platform", "github", "Platform to prove (ethereum / twitter / github / keybase)")
	flag_identity       = flag.String("identity", "username", "Identity on platform")
	flag_secret_key     = flag.String("sk", "", "Secret key of persona (without 0x)")
	flag_eth_secret_key = flag.String("eth-sk", "", "Secret key of eth wallet (if platform = ethereum)")
	flag_previous       = flag.String("previous", "", "Previous proof chain signature (Base64)")
	flag_action         = flag.String("action", "create", "Action (create / delete)")
	flag_operation      = flag.String("operation", "create", "Action (generate/ create / delete / query)")

	post_regex = regexp.MustCompile("%SIG_BASE64%")
)

func init_validators() {
	twitter.Init()
	github.Init()
	keybase.Init()
	ethereum.Init()
}

func get_action() types.Action {
	switch *flag_action {
	case "create":
		return types.Actions.Create
	case "delete":
		return types.Actions.Delete
	default:
		panic(fmt.Sprintf("Unknown action: %s", *flag_action))
	}
}

func main() {
	flag.Parse()

	switch *flag_operation {
	case "generate":
		generatePayload()
	case "query":
		query.QueryProof(*flag_platform, *flag_identity)
	}
}

func generatePayload() {
	init_validators()
	secret_key, err := ethcrypto.HexToECDSA(*flag_secret_key)
	if err != nil {
		panic(err)
	}

	platform := types.Platform(*flag_platform)
	platform_factory, ok := validator.PlatformFactories[platform]
	if !ok {
		panic(fmt.Sprintf("Platform %s not found", platform))
	}

	base := validator.Base{
		Platform:  platform,
		Previous:  *flag_previous,
		Action:    get_action(),
		Pubkey:    &secret_key.PublicKey,
		Identity:  strings.ToLower(*flag_identity),
		CreatedAt: time.Now(),
		Uuid:      uuid.New(),
	}
	validator := platform_factory(&base)

	sign_payload := validator.GenerateSignPayload()
	fmt.Printf("Sign payload: vvvvvvvvv\n%s\n^^^^^^^^^^^^^^^\n\n", sign_payload)

	if platform == types.Platforms.Ethereum {
		generate_ethereum(sign_payload, secret_key)
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

func generate_ethereum(sign_payload string, persona_sk *ecdsa.PrivateKey) {
	persona_sig, _ := crypto.SignPersonal([]byte(sign_payload), persona_sk)
	fmt.Printf("Persona sig: vvvvvvvvvv\n%s\n^^^^^^^^^^^^^^^\n\n", base64.StdEncoding.EncodeToString(persona_sig))

	wallet_sk, err := ethcrypto.HexToECDSA(*flag_eth_secret_key)
	if err != nil {
		panic(fmt.Sprintf("ETH secret key failed: %s", err.Error()))
	}
	wallet_sig, _ := crypto.SignPersonal([]byte(sign_payload), wallet_sk)
	fmt.Printf("Wallet sig: vvvvvvvvvv\n%s\n^^^^^^^^^^^^^^^\n\n", base64.StdEncoding.EncodeToString(wallet_sig))

	persona_pk := &persona_sk.PublicKey

	req := controller.ProofUploadRequest{
		Action:    types.Action(*flag_action),
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
