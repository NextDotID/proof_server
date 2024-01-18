package main

import (
	"encoding/base64"
	"fmt"

	"github.com/nextdotid/proof_server/util/crypto"
)

func main() {
	payload := "{\"action\":\"create\",\"created_at\":\"1705413052\",\"identity\":\"johndic94329223\",\"platform\":\"twitter\",\"prev\":null,\"uuid\":\"6f1fa69c-13f8-4bec-acf6-c24003633df8\"}"
	signatureBase64 := "VUwqMOFtkGGu0RqIy2HhoOyWZIdNEd29IP5ESeaWIMcZjAyCn3t1/0CmN5WaISTi1RFUOVCSw9WKC3mh78YKihw="
	signature, err := base64.StdEncoding.DecodeString(signatureBase64)
	if err != nil {
		panic(err)
	}
	pubkey, err := crypto.RecoverPubkeyFromPersonalSignature(payload, signature)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Public key: 0x%s\n", crypto.CompressedPubkeyHex(pubkey))
	fmt.Println("Success")
}
