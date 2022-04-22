package generate

import (
	"bufio"
	"crypto/ecdsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/go-resty/resty/v2"
	"github.com/nextdotid/proof-server/config"
	"github.com/nextdotid/proof-server/controller"
	"github.com/nextdotid/proof-server/types"
	"github.com/nextdotid/proof-server/util/base1024"
	"github.com/nextdotid/proof-server/util/crypto"
	"net/http"
	"os"
	"regexp"
	"strconv"
)

var post_regex = regexp.MustCompile("%SIG_BASE64%")

type GenerateParams struct {
	Platform           string
	Action             string
	PersonaPrivateKey  *ecdsa.PrivateKey
	EthereumPrivateKey *ecdsa.PrivateKey
	Identity           string
}

func GeneratePayload() {
	config.InitCliConfig()
	params := initParams()

	url := getPayloadUrl()
	personaPublicKey := &params.PersonaPrivateKey.PublicKey
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
	if err != nil || resp.StatusCode() != http.StatusOK {
		panic(fmt.Sprintf("fail to get the response resp:%v err:%v", resp, err))
	}

	err = json.Unmarshal(resp.Body(), &respPayload)
	if err != nil {
		panic(fmt.Sprintf("Unmarshal Payload Response Error, err:%v", err))
	}

	var signature, walletSignature []byte

	signature, err = crypto.SignPersonal([]byte(respPayload.SignPayload), params.PersonaPrivateKey)
	if err != nil {
		panic(fmt.Sprintf("SignPayload Error, err:%v", err))
	}

	if types.Platform(params.Platform) == types.Platforms.Ethereum {
		fmt.Printf("Post base64 encode payload: vvvvvvvvvv\n%s\n^^^^^^^^^^^^^^^\n\n", base64.StdEncoding.EncodeToString(signature))
		walletSignature, _ = crypto.SignPersonal([]byte(respPayload.SignPayload), params.EthereumPrivateKey)
		fmt.Printf("Wallet base64 sig: vvvvvvvvvv\n%s\n^^^^^^^^^^^^^^^\n\n", base64.StdEncoding.EncodeToString(walletSignature))

		fmt.Printf("Post base1024 encode payload: vvvvvvvvvv\n%s\n^^^^^^^^^^^^^^^\n\n", base1024.EncodeToString(signature))
		fmt.Printf("Wallet base1024 sig: vvvvvvvvvv\n%s\n^^^^^^^^^^^^^^^\n\n", base1024.EncodeToString(walletSignature))
	} else {
		for lang_code, payload := range respPayload.PostContent {
			fmt.Printf(
				"Post base64 encode payload [%s]: vvvvvvv\n%s\n^^^^^^^^^^\n\n",
				lang_code,
				string(post_regex.ReplaceAll([]byte(payload), []byte(base64.StdEncoding.EncodeToString(signature)))),
			)

			fmt.Printf(
				"Post base1024 encode payload [%s]: vvvvvvv\n%s\n^^^^^^^^^^\n\n",
				lang_code,
				string(post_regex.ReplaceAll([]byte(payload), []byte(base1024.EncodeToString(signature)))),
			)
		}
	}

	fmt.Printf("Need to upload the proof?\n 1. yes\n 2. no\n Press the number:\n")
	input := bufio.NewScanner(os.Stdin)
	input.Scan()
	nextStep, _ := strconv.Atoi(input.Text())

	if nextStep != 1 {
		fmt.Println("no need to continue...")
		os.Exit(0)
	}

	UploadToProof(params, personaPublicKeyParams, respPayload.CreatedAt, respPayload.Uuid, signature, walletSignature)
}

func initParams() GenerateParams {
	input := bufio.NewScanner(os.Stdin)
	fmt.Println("For the generate signature process, need your Persona Private Key at first step.Please enter your Persona Private Key (without 0x prefix):")
	input.Scan()
	pk := input.Text()
	personaPrivateKey, err := ethcrypto.HexToECDSA(pk)
	if err != nil {
		panic(fmt.Sprintf("Get Persona PrivateKey Error, err:%v", err))
	}

	fmt.Println("\nThe following facts also need to use in signature generation process")
	fmt.Println("Platform (find out the support platform at README.md):")
	input.Scan()
	platform := input.Text()
	fmt.Println("\nIdentity (find out the identity of each platform at README.md):")
	input.Scan()
	identity := input.Text()

	fmt.Println("\nAction (create or delete):")
	input.Scan()
	action := input.Text()

	gp := GenerateParams{
		Platform:          platform,
		Identity:          identity,
		Action:            action,
		PersonaPrivateKey: personaPrivateKey,
	}

	if types.Platform(platform) == types.Platforms.Ethereum {
		fmt.Println("\nEthereum Private Key (without 0x prefix):")
		input.Scan()
		ek := input.Text()
		ethereumPrivateKey, err := ethcrypto.HexToECDSA(ek)
		if err != nil {
			panic(fmt.Sprintf("Get Persona PrivateKey Error, err:%v", err))
		}
		gp.EthereumPrivateKey = ethereumPrivateKey
	}

	return gp
}

func getPayloadUrl() string {
	return config.Viper.GetString("server.hostname") + config.Viper.GetString("server.generate_path")
}
