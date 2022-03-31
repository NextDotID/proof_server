package generate

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/go-resty/resty/v2"
	"github.com/nextdotid/proof-server/cli"
	"github.com/nextdotid/proof-server/controller"
	"github.com/nextdotid/proof-server/types"
	"github.com/nextdotid/proof-server/util/crypto"
	"regexp"
)

var post_regex = regexp.MustCompile("%SIG_BASE64%")

type GenerateParams struct {
	Platform           string `json:"platform"`
	Previous           string `json:"previous"`
	Action             string `json:"action"`
	PersonaPrivateKey  string `json:"persona_private_key"`
	EthereumPrivateKey string `json:"ethereum_private_key"`

	Identity string `json:"identity"` // Identity on target platform.

	Signature       string `json:"signature"`
	WalletSignature string `json:"wallet_signature"`

	CreatedAt string `json:"created_at"`
	Uuid      string `json:"uuid"` // Uuid gives this link an unique identifier, to let other
}

func GeneratePayload() {
	cli.InitConfig()
	params := initParams()
	personaPrivateKey, err := ethcrypto.HexToECDSA(params.PersonaPrivateKey)
	if err != nil {
		panic(err)
	}

	url := getPayloadUrl()
	personaPublicKey := &personaPrivateKey.PublicKey
	req := controller.ProofPayloadRequest{
		Action:    types.Action(params.Action),
		Platform:  types.Platform(params.Platform),
		Identity:  params.Identity,
		PublicKey: "0x" + crypto.CompressedPubkeyHex(personaPublicKey),
	}

	client := resty.New()
	resp, err := client.R().SetBody(req).EnableTrace().Post(url)
	respPayload := controller.ProofPayloadResponse{}

	err = json.Unmarshal(resp.Body(), &respPayload)
	if err != nil {
		fmt.Printf("Unmarshal Payload Response Error, err:%v", err)
		return
	}

	signature, err := crypto.SignPersonal([]byte(respPayload.SignPayload), personaPrivateKey)
	if err != nil {
		fmt.Printf("SignPayload Error, err:%v", err)
		return
	}

	if types.Platform(params.Platform) == types.Platforms.Twitter {
		for lang_code, payload := range respPayload.PostContent {
			fmt.Printf(
				"Post payload [%s]: vvvvvvv\n%s\n^^^^^^^^^^\n\n",
				lang_code,
				string(post_regex.ReplaceAll([]byte(payload), []byte(base64.StdEncoding.EncodeToString(signature)))),
			)
		}
	} else if types.Platform(params.Platform) == types.Platforms.Ethereum {
		//persona_sig, _ := crypto.SignPersonal([]byte(sign_payload), personaPrivateKey)
		fmt.Printf("Persona sig: vvvvvvvvvv\n%s\n^^^^^^^^^^^^^^^\n\n", base64.StdEncoding.EncodeToString(signature))

		ethereumPrivateKey, err := ethcrypto.HexToECDSA(params.EthereumPrivateKey)
		if err != nil {
			panic(fmt.Sprintf("ETH secret key failed: %s", err.Error()))
		}
		walletSignature, _ := crypto.SignPersonal([]byte(respPayload.SignPayload), ethereumPrivateKey)
		fmt.Printf("Wallet sig: vvvvvvvvvv\n%s\n^^^^^^^^^^^^^^^\n\n", base64.StdEncoding.EncodeToString(walletSignature))
	}

	fmt.Printf("CreateAt time:  vvvvvvvvvv\n%s\n^^^^^^^^^^^^^^^\n\n", respPayload.CreatedAt)
	fmt.Printf("UUID:  vvvvvvvvvv\n%s\n^^^^^^^^^^^^^^^\n\n", respPayload.Uuid)
}

//func generate_ethereum(sign_payload string, persona_sk *ecdsa.PrivateKey, action string, eth_sk string) {
//	persona_sig, _ := crypto.SignPersonal([]byte(sign_payload), persona_sk)
//	fmt.Printf("Persona sig: vvvvvvvvvv\n%s\n^^^^^^^^^^^^^^^\n\n", base64.StdEncoding.EncodeToString(persona_sig))
//
//	wallet_sk, err := ethcrypto.HexToECDSA(eth_sk)
//	if err != nil {
//		panic(fmt.Sprintf("ETH secret key failed: %s", err.Error()))
//	}
//	wallet_sig, _ := crypto.SignPersonal([]byte(sign_payload), wallet_sk)
//	fmt.Printf("Wallet sig: vvvvvvvvvv\n%s\n^^^^^^^^^^^^^^^\n\n", base64.StdEncoding.EncodeToString(wallet_sig))
//
//	persona_pk := &persona_sk.PublicKey
//
//	req := controller.ProofUploadRequest{
//		Action:    types.Action(action),
//		Platform:  types.Platforms.Ethereum,
//		Identity:  ethcrypto.PubkeyToAddress(wallet_sk.PublicKey).String(),
//		PublicKey: "0x" + crypto.CompressedPubkeyHex(persona_pk),
//		Extra: controller.ProofUploadRequestExtra{
//			EthereumWalletSignature: base64.StdEncoding.EncodeToString(wallet_sig),
//			Signature:               base64.StdEncoding.EncodeToString(persona_sig),
//		},
//	}
//
//	req_json, _ := json.Marshal(req)
//	fmt.Printf("POST /v1/proof/payload request:\n\n%s\n\n", req_json)
//}

func initParams() GenerateParams {
	return GenerateParams{
		Platform:           cli.Viper.GetString("cli.params.platform"),
		Identity:           cli.Viper.GetString("cli.params.identity"),
		Previous:           cli.Viper.GetString("cli.params.previous"),
		Action:             cli.Viper.GetString("cli.params.action"),
		PersonaPrivateKey:  cli.Viper.GetString("cli.params.persona_private_key"),
		EthereumPrivateKey: cli.Viper.GetString("cli.params.ethereum_private_key"),
	}
}

func getPayloadUrl() string {
	return cli.Viper.GetString("server.hostname") + cli.Viper.GetString("server.generate_path")
}
