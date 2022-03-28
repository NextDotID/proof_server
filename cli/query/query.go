package query

import (
	"fmt"
	"github.com/go-resty/resty/v2"
)

func QueryProof(url string, platform string, identity string) {
	rp := make(map[string]string)
	rp["platform"] = platform
	rp["identity"] = identity

	client := resty.New()
	resp, _ := client.R().SetQueryParams(rp).EnableTrace().Get(url)

	fmt.Println(resp)
}
