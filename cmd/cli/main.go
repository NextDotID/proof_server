package main

import (
	"flag"
	"fmt"
	"github.com/nextdotid/proof-server/cli/generate"
	"github.com/nextdotid/proof-server/cli/query"
	"github.com/nextdotid/proof-server/cli/upload"
	"github.com/nextdotid/proof-server/config"
	"github.com/nextdotid/proof-server/controller"
	"github.com/nextdotid/proof-server/types"
	"regexp"
)

var (
	flag_platform       = flag.String("platform", "github", "Platform to prove (ethereum / twitter / github / keybase)")
	flag_identity       = flag.String("identity", "username", "Identity on platform")
	flag_secret_key     = flag.String("sk", "", "Secret key of persona (without 0x)")
	flag_eth_secret_key = flag.String("eth-sk", "", "Secret key of eth wallet (if platform = ethereum)")
	flag_previous       = flag.String("previous", "", "Previous proof chain signature (Base64)")
	flag_action         = flag.String("action", "create", "Action (create / delete)")
	flag_operation      = flag.String("operation", "create", "Action (generate/ create / delete / query)")
	flag_public_key     = flag.String("public-key", "", "Public key of persona")

	post_regex = regexp.MustCompile("%SIG_BASE64%")
)

func main() {
	flag.Parse()
	cfg := config.GetConfigOfCli("./config/config.cli.json")

	switch *flag_operation {
	case "generate":
		generate.GeneratePayload(*flag_secret_key, *flag_platform, *flag_identity, *flag_previous, *flag_action, *flag_eth_secret_key)
	case "query":
		url := cfg.ServerURL + cfg.QueryPath
		query.QueryProof(url, *flag_platform, *flag_identity)
	case "upload":
		url := cfg.ServerURL + cfg.UploadPath
		req := controller.ProofUploadRequest{}
		req.Platform = types.Platform(*flag_platform)
		req.Identity = *flag_identity
		req.PublicKey = *flag_public_key
		req.Action = types.Action(*flag_action)
		req.Uuid = ""
		upload.UploadToProof(url, req)
	default:
		fmt.Printf("Unknow Operation: %s", *flag_operation)
	}
}
