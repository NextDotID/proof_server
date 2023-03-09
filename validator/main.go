package validator

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-faster/errors"
	"github.com/google/uuid"
	"github.com/nextdotid/proof_server/config"
	"github.com/nextdotid/proof_server/headless"
	"github.com/nextdotid/proof_server/types"
	"github.com/samber/lo"
)

var (
	// PlatformFactories contains all supported platform factory.
	PlatformFactories map[types.Platform]func(*Base) IValidator
)

type IValidator interface {
	// GeneratePostPayload gives a post structure (with
	// placeholders) for user to post on target platform.
	GeneratePostPayload() (post map[string]string)
	// GenerateSignPayload generates a string to be signed.
	GenerateSignPayload() (payload string)

	Validate() (err error)
}

type Base struct {
	Platform types.Platform
	Previous string
	Action   types.Action
	Pubkey   *ecdsa.PublicKey
	// Identity on target platform.
	Identity         string
	AltID            string
	ProofLocation    string
	Signature        []byte
	SignaturePayload string
	Text             string
	// Extra info needed by separate platforms (e.g. Ethereum)
	Extra map[string]string
	// CreatedAt indicates creation time of this link
	CreatedAt time.Time
	// Uuid gives this link an unique identifier, to let other
	// third-party service distinguish / store / dedup links with
	// ease.
	Uuid uuid.UUID
}

// BaseToInterface converts a `validator.Base` struct to
// `validator.IValidator` interface.
func BaseToInterface(v *Base) IValidator {
	performer_factory, ok := PlatformFactories[v.Platform]
	if !ok {
		return nil
	}

	return performer_factory(v)
}

func GetPostWithHeadlessBrowser(url string, selector string, regexp string, property string) (post string, err error) {
	headlessEntrypoint := lo.Sample(config.C.Headless.Urls)
	headlessEntrypoint += "/v1/find"
	request := headless.FindRequest{
		Location: url,
		Timeout:  "120s",
		Match: headless.Match{
			Type: "regexp",
			MatchRegExp: &headless.MatchRegExp{
				Selector: selector,
				Value:    regexp,
			},
			MatchXPath: nil,
			MatchJS:    nil,
			Property:   property,
		},
	}
	// POST request body to entrypoint headless server
	requestBody, err := json.Marshal(request)
	if err != nil {
		return "", err
	}
	respRaw, err := http.Post(headlessEntrypoint, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}
	defer respRaw.Body.Close()
	response := headless.FindRespond{}
	err = json.NewDecoder(respRaw.Body).Decode(&response)
	if err != nil {
		return "", err
	}
	if response.Message != "" {
		return "", errors.Errorf("Error when fetching post from headless browser: %s", response.Message)
	}

	return response.Content, nil
}

// H for JSON builder.
type H map[string]any
