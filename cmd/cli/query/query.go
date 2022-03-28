package query

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/nextdotid/proof-server/controller"
	"strings"
)

func QueryProof(platform string, identity string) {
	req := controller.ProofQueryRequest{}
	req.Platform = platform
	req.Identity = strings.Split(req.Identity[0], ",")

	client := resty.New()
	url := fmt.Sprintf("http://localhost:9800/v1/proof?identity=%s&platform=%s", identity, platform)
	resp, _ := client.R().EnableTrace().Get(url)

	fmt.Println(resp)
}
