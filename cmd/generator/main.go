package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"regexp"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/nextdotid/proof-server/types"
	"github.com/nextdotid/proof-server/util/crypto"
	"github.com/nextdotid/proof-server/validator"
	"github.com/nextdotid/proof-server/validator/github"
	"github.com/nextdotid/proof-server/validator/keybase"
	"github.com/nextdotid/proof-server/validator/twitter"
)

var (
	flag_platform = flag.String("platform", "github", "Platform to prove")
	flag_identity = flag.String("identity", "username", "Identity on platform")
	flag_secret_key = flag.String("sk", "", "Secret key of persona (without 0x)")
	flag_previous = flag.String("previous", "", "Previous proof chain signature (Base64)")
	flag_action = flag.String("action", "create", "Action (create / delete)")

	post_regex = regexp.MustCompile("%%SIG_BASE64%%")
)

func init_validators() {
	twitter.Init()
	github.Init()
	keybase.Init()
	// TODO: support ethereum
	// ethereum.Init()
}

func get_action() types.Action{
	switch *flag_action {
	case "create":
		return types.Actions.Create
	case "delete":
		return types.Actions.Delete
	default:
		panic(fmt.Sprintf("Unknown action: %s", *flag_action))
	}
}

func main() {
	flag.Parse()
	init_validators()
	secret_key, err := ethcrypto.HexToECDSA(*flag_secret_key)
	if err != nil {
		panic(err)
	}

	platform := types.Platform(*flag_platform)
	platform_factory, ok := validator.PlatformFactories[platform]
	if !ok {
		panic(fmt.Sprintf("Platform %s not found", platform))
	}

	validator := platform_factory(validator.Base{
		Platform:      platform,
		Previous:      *flag_previous,
		Action:        get_action(),
		Pubkey:        &secret_key.PublicKey,
		Identity:      *flag_identity,
	})

	sign_payload := validator.GenerateSignPayload()
	fmt.Printf("Sign payload: vvvvvvvvv\n%s\n^^^^^^^^^^^^^^^\n\n", sign_payload)

	raw_post := validator.GeneratePostPayload()
	signature, err := crypto.SignPersonal([]byte(sign_payload), secret_key)
	if err != nil {
		panic(err)
	}
	post := post_regex.ReplaceAll([]byte(raw_post), []byte(base64.StdEncoding.EncodeToString(signature)))
	fmt.Printf("Post payload: vvvvvvvvv\n%s\n^^^^^^^^^^^^^^^\n\n", post)
}
