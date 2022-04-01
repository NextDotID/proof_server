package generate

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/go-resty/resty/v2"
	"github.com/nextdotid/proof-server/config"
	"github.com/nextdotid/proof-server/controller"
	"github.com/nextdotid/proof-server/types"
	"github.com/nextdotid/proof-server/util/crypto"
	"github.com/spf13/cast"
	"os"
	"regexp"
)

var post_regex = regexp.MustCompile("%SIG_BASE64%")

type GenerateParams struct {
	Platform           string
	Action             string
	PersonaPrivateKey  string
	EthereumPrivateKey string
	Identity           string
}

func GeneratePayload() {
	config.InitCliConfig()
	params := initParams()

	personaPrivateKey, err := ethcrypto.HexToECDSA(params.PersonaPrivateKey)
	if err != nil {
		fmt.Printf("Get Persona PrivateKey Error, err:%v", err)
		return
	}

	url := getPayloadUrl()
	personaPublicKey := &personaPrivateKey.PublicKey
	personaPublicKeyParams := "0x" + crypto.CompressedPubkeyHex(personaPublicKey)
	req := controller.ProofPayloadRequest{
		Action:    types.Action(params.Action),
		Platform:  types.Platform(params.Platform),
		Identity:  params.Identity,
		PublicKey: personaPublicKeyParams,
	}

	client := resty.New()
	resp, err := client.R().SetBody(req).EnableTrace().Post(url)
	respPayload := controller.ProofPayloadResponse{}

	err = json.Unmarshal(resp.Body(), &respPayload)
	if err != nil {
		fmt.Printf("Unmarshal Payload Response Error, err:%v", err)
		return
	}

	var signature, walletSignature []byte

	signature, err = crypto.SignPersonal([]byte(respPayload.SignPayload), personaPrivateKey)
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
		walletSignature, _ = crypto.SignPersonal([]byte(respPayload.SignPayload), ethereumPrivateKey)
		fmt.Printf("Wallet sig: vvvvvvvvvv\n%s\n^^^^^^^^^^^^^^^\n\n", base64.StdEncoding.EncodeToString(walletSignature))
	} else {
		fmt.Printf("Persona sig: vvvvvvvvvv\n%s\n^^^^^^^^^^^^^^^\n\n", base64.StdEncoding.EncodeToString(signature))
	}

	fmt.Printf("Need to upload the proof?\n 1. yes\n 2. no\n Press the number:\n")
	input := bufio.NewScanner(os.Stdin)
	input.Scan()
	nextStep := cast.ToInt(input.Text())

	if nextStep != 1 {
		fmt.Println("no need to continue...")
		os.Exit(0)
	}

	UploadToProof(params, personaPublicKeyParams, respPayload.CreatedAt, respPayload.Uuid, signature, walletSignature)
}

func initParams() GenerateParams {
	input := bufio.NewScanner(os.Stdin)
	fmt.Println("For the generate signature process, need your Persona Private Key at first step, Persona Private Key:")
	input.Scan()
	pk := input.Text()

	fmt.Println("\nThe following facts also need to generate signature process")
	fmt.Println("Platform (find out a support platform at README.md):")
	input.Scan()
	platform := input.Text()
	fmt.Println("\nIdentity (find out the identity of each platform at README.md):")
	input.Scan()
	identity := input.Text()

	fmt.Println("\nAction (create or delete):")
	input.Scan()
	action := input.Text()

	ek := ""
	if types.Platform(platform) == types.Platforms.Ethereum {
		fmt.Println("\nEthereum Private Key:")
		input.Scan()
		ek = input.Text()
	}

	return GenerateParams{
		Platform:           platform,
		Identity:           identity,
		Action:             action,
		PersonaPrivateKey:  pk,
		EthereumPrivateKey: ek,
	}
}

func getPayloadUrl() string {
	return config.Viper.GetString("server.hostname") + config.Viper.GetString("server.generate_path")
}
