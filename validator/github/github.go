package github

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/nextdotid/proof-server/types"
	"github.com/nextdotid/proof-server/util"
	"github.com/nextdotid/proof-server/util/crypto"
	"github.com/nextdotid/proof-server/validator"
	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"

	ghub "github.com/google/go-github/v41/github"
)

type Github struct {
	*validator.Base
}

type gistPayload struct {
	Version        string `json:"version"`
	Comment        string `json:"comment"`
	Comment2       string `json:"comment2"`
	Persona        string `json:"persona"`
	GithubUsername string `json:"github_username"`
	SignPayload    string `json:"sign_payload"`
	Signature      string `json:"signature"`
	CreatedAt      string `json:"created_at"`
	Uuid           string `json:"uuid"`
}

var (
	l = logrus.WithFields(logrus.Fields{"module": "validator", "validator": "github"})
)

func Init() {
	if validator.PlatformFactories == nil {
		validator.PlatformFactories = make(map[types.Platform]func(*validator.Base) validator.IValidator)
	}
	validator.PlatformFactories[types.Platforms.Github] = func(base *validator.Base) validator.IValidator {
		gh := Github{base}
		return &gh
	}
}

func (gh *Github) GeneratePostPayload() (post map[string]string) {
	gh.Identity = strings.ToLower(gh.Identity)
	payload := gistPayload{
		Version:        "1",
		Comment:        "Here's an NextID proof of this Github account.",
		Comment2:       "To validate, base64.decode the signature, and recover pubkey from it using sign_payload with ethereum personal_sign algo.",
		Persona:        "0x" + crypto.CompressedPubkeyHex(gh.Pubkey),
		GithubUsername: gh.Identity,
		SignPayload:    gh.GenerateSignPayload(),
		Signature:      "%SIG_BASE64%",
		CreatedAt:      util.TimeToTimestampString(gh.CreatedAt),
		Uuid:           gh.Uuid.String(),
	}

	payload_json, _ := json.MarshalIndent(payload, "", "\t")
	return map[string]string{"default": string(payload_json)}
}

func (gh *Github) GenerateSignPayload() (payload string) {
	gh.Identity = strings.ToLower(gh.Identity)
	payloadStruct := validator.H{
		"action":     string(gh.Action),
		"identity":   gh.Identity,
		"platform":   string(types.Platforms.Github),
		"prev":       nil,
		"created_at": util.TimeToTimestampString(gh.CreatedAt),
		"uuid":       gh.Uuid.String(),
	}
	if gh.Previous != "" {
		payloadStruct["prev"] = gh.Previous
	}

	payload_bytes, _ := json.Marshal(payloadStruct)
	return string(payload_bytes)
}

func (gh *Github) Validate() (err error) {
	gh.Identity = strings.ToLower(gh.Identity)
	gh.SignaturePayload = gh.GenerateSignPayload()

	client := ghub.NewClient(nil)
	gist, response, err := client.Gists.Get(context.TODO(), gh.ProofLocation)
	if err != nil {
		return xerrors.Errorf("error when fetching gist: %w", err)
	}

	if response.StatusCode != 200 {
		return xerrors.Errorf("error when fetching gist")
	}

	if gh.Identity != gist.Owner.GetLogin() {
		return xerrors.Errorf("gist owner mismatch: should be %s, but got %s", gh.Identity, gist.Owner.GetLogin())
	}

	gist_filename := fmt.Sprintf("0x%s.json", crypto.CompressedPubkeyHex(gh.Pubkey))
	files := gist.GetFiles()
	content := ""
	for filename, file := range files {
		if filename != ghub.GistFilename(gist_filename) {
			continue
		}

		content = *file.Content
	}
	if content == "" {
		return xerrors.Errorf("%s not found or empty", gist_filename)
	}
	payload := gistPayload{}
	err = json.Unmarshal([]byte(content), &payload)
	if err != nil {
		return xerrors.Errorf("error when parsing JSON: %w", err)
	}

	pubkey_recovered, err := crypto.StringToPubkey(payload.Persona)
	if err != nil {
		return xerrors.Errorf("error when recovering pubkey: %w", err)
	}
	signature, err := util.DecodeString(payload.Signature)
	if err != nil {
		return xerrors.Errorf("error when decoding signature: %w", err)
	}
	return crypto.ValidatePersonalSignature(payload.SignPayload, signature, pubkey_recovered)
}
