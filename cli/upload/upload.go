package upload

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/nextdotid/proof-server/cli"
	"github.com/nextdotid/proof-server/controller"
	"github.com/nextdotid/proof-server/types"
	"strings"
)

type UploadParams struct {
	Platform                string `json:"platform"`
	Action                  string `json:"action"`
	PublicKey               string `json:"public_key"`
	Identity                string `json:"identity"` // Identity on target platform
	PersonaSignature        string `json:"signature"`
	EthereumWalletSignature string `json:"ethereum_wallet_signature"`
	CreatedAt               string `json:"created_at"`
	Uuid                    string `json:"uuid"` // Uuid gives this link an unique identifier, to let other
	ProofLocation           string `json:"proof_location"`
}

func UploadToProof() {
	cli.InitConfig()
	params := initParams()
	req := controller.ProofUploadRequest{
		Action:        types.Action(params.Action),
		Platform:      types.Platform(params.Platform),
		Identity:      strings.ToLower(params.Identity),
		PublicKey:     params.PublicKey,
		CreatedAt:     params.CreatedAt,
		Uuid:          params.Uuid,
		ProofLocation: params.ProofLocation,
	}
	req.Extra.Signature = params.PersonaSignature

	if types.Platform(params.Platform) == types.Platforms.Ethereum {
		req.Extra.EthereumWalletSignature = params.EthereumWalletSignature
	}

	url := getUploadUrl()
	client := resty.New()
	resp, err := client.R().SetBody(req).EnableTrace().Post(url)

	fmt.Println(resp)
	fmt.Println(err)
}

func getUploadUrl() string {
	return cli.Viper.GetString("server.hostname") + cli.Viper.GetString("server.upload_path")
}

func initParams() UploadParams {
	return UploadParams{
		Platform:                cli.Viper.GetString("cli.params.platform"),
		Identity:                cli.Viper.GetString("cli.params.identity"),
		Action:                  cli.Viper.GetString("cli.params.action"),
		PersonaSignature:        cli.Viper.GetString("cli.params.persona_signature"),
		EthereumWalletSignature: cli.Viper.GetString("cli.params.ethereum_wallet_signature"),
		CreatedAt:               cli.Viper.GetString("cli.params.create_at"),
		Uuid:                    cli.Viper.GetString("cli.params.uuid"),
		ProofLocation:           cli.Viper.GetString("cli.params.proof_location"),
		PublicKey:               cli.Viper.GetString("cli.params.public_key"),
	}
}
