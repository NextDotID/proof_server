package main

import (
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/nextdotid/proof_server/config"
	"github.com/nextdotid/proof_server/types"
	"github.com/nextdotid/proof_server/util"
	"github.com/nextdotid/proof_server/util/crypto"
	mycrypto "github.com/nextdotid/proof_server/util/crypto"
	"github.com/nextdotid/proof_server/validator"
	"github.com/nextdotid/proof_server/validator/twitter"
)

func generate() twitter.Twitter {
	pubkey, _ := mycrypto.StringToSecp256k1Pubkey("0x02492e9cb3a3578acc27fd1884a6de1758add291300754557d06a28308951d46ea")
	created_at, _ := util.TimestampStringToTime("1697953689")
	return twitter.Twitter{
		Base: &validator.Base{
			Platform:      types.Platforms.Twitter,
			Previous:      "",
			Action:        types.Actions.Create,
			Pubkey:        pubkey,
			Identity:      "askcasmir",
			ProofLocation: "1715976641241919493",
			Text:          "",
			Uuid:          uuid.MustParse("64442f89-9cd9-4f62-bbd0-47e5f849f9b4"),
			CreatedAt:     created_at,
		},
	}
}

func main() {
	config.Init("./config/config.json")

	myTwitter := generate()
	sigBytes, err := util.DecodeString("Upf+OxdAzaVb0mVxso0PlTDQYf6JjldY/xEo7RkCMM9dM7IgBGgWU5Yk5U5j0RdhmX64Y9GCqziyQD9tHIxxFRs=")
	if err != nil {
		panic(err)
	}

	myTwitter.SignaturePayload = myTwitter.GenerateSignPayload()
	err = crypto.ValidatePersonalSignature(myTwitter.SignaturePayload, sigBytes, myTwitter.Pubkey)

	// err := myTwitter.Validate()
	if err != nil {
		panic(err)
	}
	fmt.Println("Validated")
	os.Exit(0)
}
