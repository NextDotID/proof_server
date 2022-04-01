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

func UploadToProof(gp GenerateParams, ppk string, createAt string, uuid string, sg []byte, wg []byte) {
	config.InitCliConfig()
	var pl string
	if types.Platform(gp.Platform) != types.Platforms.Ethereum {
		input := bufio.NewScanner(os.Stdin)
		fmt.Println("Proof Location (find out how to get the proof location at README.md)::")
		input.Scan()
		pl = input.Text()
	}

	req := controller.ProofUploadRequest{
		Action:        types.Action(gp.Action),
		Platform:      types.Platform(gp.Platform),
		Identity:      strings.ToLower(gp.Identity),
		PublicKey:     ppk,
		CreatedAt:     createAt,
		Uuid:          uuid,
		ProofLocation: pl,
	}

	req.Extra.Signature = base64.StdEncoding.EncodeToString((sg))
	if types.Platform(gp.Platform) == types.Platforms.Ethereum {
		req.Extra.EthereumWalletSignature = base64.StdEncoding.EncodeToString((wg))
	}

	url := getUploadUrl()
	client := resty.New()
	resp, err := client.R().SetBody(req).EnableTrace().Post(url)

	if resp.StatusCode() == http.StatusCreated {
		fmt.Println("Upload succeed!!")
	} else {
		fmt.Printf("Oops, some error occured err:%v", err)
	}

}

func getUploadUrl() string {
	return config.Viper.GetString("server.hostname") + config.Viper.GetString("server.upload_path")
}
