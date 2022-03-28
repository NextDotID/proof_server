package upload

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/nextdotid/proof-server/controller"
)

func UploadToProof(url string, req controller.ProofUploadRequest) {
	client := resty.New()
	resp, _ := client.R().SetBody(req).EnableTrace().Post(url)
	fmt.Println(resp)
}
