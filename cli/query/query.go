package query

import (
	"bufio"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/nextdotid/proof-server/config"
	"github.com/spf13/cast"
	"os"
)

type QueryParams struct {
	Platform string `json:"platform"`
	Identity string `json:"identity"` // Identity on target platform.
	Page     int    `json:"page"`
}

func QueryProof() {
	config.InitCliConfig()
	params := initParams()
	req := make(map[string]string)
	req["platform"] = params.Platform
	req["identity"] = params.Identity
	req["page"] = cast.ToString(params.Page)

	url := getQueryUrl()
	client := resty.New()
	resp, err := client.R().SetQueryParams(req).EnableTrace().Get(url)

	fmt.Println(resp)
	fmt.Println(err)
}

func initParams() QueryParams {
	input := bufio.NewScanner(os.Stdin)
	fmt.Println("For the query process, we could have platform/identity/page as the query condition\n")
	fmt.Println("Platform (find out a support platform at README.md):")
	input.Scan()
	platform := input.Text()
	fmt.Println("\nIdentity (find out the identity of each platform at README.md):")
	input.Scan()
	identity := input.Text()
	fmt.Println("\nPage (We will give maximum 20 results for each query, you can give a page number for getting more results):")
	input.Scan()
	page := input.Text()

	return QueryParams{
		Platform: platform,
		Identity: identity,
		Page:     cast.ToInt(page),
	}
}

func getQueryUrl() string {
	return config.Viper.GetString("server.hostname") + config.Viper.GetString("server.query_path")
}
