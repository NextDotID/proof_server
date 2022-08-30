package query

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/go-resty/resty/v2"
	"github.com/nextdotid/proof_server/config"
)

type QueryParams struct {
	Platform string `json:"platform"`
	Identity string `json:"identity"` // Identity on target platform.
	Page     string `json:"page"`
}

func QueryProof() {
	config.InitCliConfig()
	params := initParams()
	req := make(map[string]string)
	req["platform"] = params.Platform
	req["identity"] = params.Identity
	req["page"] = params.Page
	getAndPrintData(req)

	input := bufio.NewScanner(os.Stdin)
	for {
		fmt.Println("\nChoose next step:\n 1. next page\n 2. Get the data of the specified page\n 3. Quit\n Enter the number:")
		input.Scan()
		nextStep := input.Text()
		switch nextStep {
		case "1":
			page, _ := strconv.Atoi(params.Page)
			page += 1
			fmt.Printf("\nGet the data of Page %d\n", page)
			req["page"] = strconv.FormatInt(int64(page), 10)
			getAndPrintData(req)
		case "2":
			fmt.Println("\nPlease enter the page number")
			input.Scan()
			page := input.Text()
			req["page"] = page
			getAndPrintData(req)
		case "3":
			os.Exit(0)
		default:
			panic(fmt.Sprintf("Unknown Operation %s", nextStep))
		}
	}
}

func getAndPrintData(req map[string]string) {
	url := getQueryUrl()
	client := resty.New()
	resp, err := client.R().SetQueryParams(req).EnableTrace().Get(url)
	if err != nil {
		panic(fmt.Sprintf("Oops, fail to get the result, err:%v", err))
	}
	fmt.Println(PrettyString(string(resp.Body())))
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
		Page:     page,
	}
}

func getQueryUrl() string {
	return config.Viper.GetString("server.hostname") + config.Viper.GetString("server.query_path")
}

func PrettyString(str string) (string, error) {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, []byte(str), "", "    "); err != nil {
		return "", err
	}
	return prettyJSON.String(), nil
}
