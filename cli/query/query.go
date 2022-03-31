package query

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/nextdotid/proof-server/cli"
	"github.com/spf13/cast"
)

type QueryParams struct {
	Platform string `json:"platform"`
	Identity string `json:"identity"` // Identity on target platform.
	Page     int    `json:"page"`
}

func QueryProof() {
	cli.InitConfig()
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
	return QueryParams{
		Platform: cli.Viper.GetString("cli.params.platform"),
		Identity: cli.Viper.GetString("cli.params.identity"),
		Page:     cli.Viper.GetInt("cli.params.page"),
	}
}

func getQueryUrl() string {
	return cli.Viper.GetString("server.hostname") + cli.Viper.GetString("server.query_path")
}
