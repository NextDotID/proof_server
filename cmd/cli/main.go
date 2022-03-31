package main

import (
	"flag"
	"fmt"
	"github.com/nextdotid/proof-server/cli/generate"
	"github.com/nextdotid/proof-server/cli/query"
	"github.com/nextdotid/proof-server/cli/upload"
)

var flag_operation = flag.String("operation", "query", "operation (query / generate / upload)")

func main() {
	flag.Parse()
	switch *flag_operation {
	case "generate":
		generate.GeneratePayload()
	case "query":
		query.QueryProof()
	case "upload":
		upload.UploadToProof()
	default:
		fmt.Printf("Unknow Operation: %s", *flag_operation)
	}
}
