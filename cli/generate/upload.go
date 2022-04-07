package generate

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/nextdotid/proof-server/config"
	"github.com/nextdotid/proof-server/controller"
	"github.com/nextdotid/proof-server/types"
	"net/http"
	"os"
	"strings"
)

func UploadToProof(gp GenerateParams, personaPublicKey string, createAt string, uuid string, signature []byte, walletSignature []byte) {
	config.InitCliConfig()
	var pl string
	if types.Platform(gp.Platform) != types.Platforms.Ethereum {
		input := bufio.NewScanner(os.Stdin)
		fmt.Println("Proof Location (find out how to get the proof location for each platform at README.md):")
		input.Scan()
		pl = input.Text()
	}

	req := controller.ProofUploadRequest{
		Action:        types.Action(gp.Action),
		Platform:      types.Platform(gp.Platform),
		Identity:      strings.ToLower(gp.Identity),
		PublicKey:     personaPublicKey,
		CreatedAt:     createAt,
		Uuid:          uuid,
		ProofLocation: pl,
	}

	req.Extra.Signature = base64.StdEncoding.EncodeToString((signature))
	if types.Platform(gp.Platform) == types.Platforms.Ethereum {
		req.Extra.EthereumWalletSignature = base64.StdEncoding.EncodeToString((walletSignature))
	}

	url := getUploadUrl()
	client := resty.New()
	resp, err := client.R().SetBody(req).EnableTrace().Post(url)

	if resp.StatusCode() == http.StatusCreated {
		fmt.Println("Upload succeed!!")
	} else {
		panic(fmt.Sprintf("Oops, some error occured. resp:%v err:%v", resp, err))
	}
	os.Exit(0)
}

func getUploadUrl() string {
	return config.Viper.GetString("server.hostname") + config.Viper.GetString("server.upload_path")
}
