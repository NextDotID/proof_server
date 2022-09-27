package minds

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/nextdotid/proof_server/types"
	"github.com/nextdotid/proof_server/util"
	mycrypto "github.com/nextdotid/proof_server/util/crypto"
	"github.com/nextdotid/proof_server/validator"
	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"
)

const (
	URL            = "https://www.minds.com/api/v2/entities/?urns=urn%%3Aactivity%%3A%s&as_activities=0&export_user_counts=false"
	MATCH_TEMPLATE = "^Sig: (.*)$"
)

var (
	l           = logrus.WithFields(logrus.Fields{"module": "validator", "validator": "minds"})
	re          = regexp.MustCompile(MATCH_TEMPLATE)
	POST_STRUCT = map[string]string{
		"default": "🎭 Verifying my Minds ID @%s for NextID.\n\nSig: %%SIG_BASE64%%\nCreatedAt: %d\nUUID: %s%s\n\nPowered by Next.ID - Connect All Digital Identities.\n",
	}
)

type Minds struct {
	*validator.Base
}

type MindsPayload struct {
	Status   string        `json:"status"`
	Entities []MindsEntity `json:"entities"`
}

type MindsEntity struct {
	// Guid is post ID.
	Guid string `json:"guid"`
	// TimeCreated is second-based timestamp.
	TimeCreated string `json:"time_created"`
	// TimeUpdated is second-based timestamp.
	TimeUpdated string `json:"time_updated"`
	// Message is post content.
	Message string     `json:"message"`
	Owner   MindsOwner `json:"ownerObj"`
}

type MindsOwner struct {
	// Guid is User ID
	Guid string `json:"guid"`
	// TimeCreated is second-based timestamp.
	TimeCreated string `json:"time_created"`
	// TimeUpdated is second-based timestamp.
	// TimeUpdated      string `json:"time_updated"`
	UserName         string `json:"username"`
	Name             string `json:"name"`
	BriefDescription string `json:"briefdescription"`
}

func Init() {
	if validator.PlatformFactories == nil {
		validator.PlatformFactories = make(map[types.Platform]func(*validator.Base) validator.IValidator)
	}
	validator.PlatformFactories[types.Platforms.Minds] = func(base *validator.Base) validator.IValidator {
		minds := Minds{base}
		return &minds
	}
}

func (minds *Minds) GeneratePostPayload() (post map[string]string) {
	post = make(map[string]string, 0)
	previous := ""
	if minds.Previous != "" {
		previous = "\nPrevious: " + minds.Previous
	}
	for lang_code, template := range POST_STRUCT {
		post[lang_code] = fmt.Sprintf(template, minds.Identity, minds.CreatedAt.Unix(), minds.Uuid.String(), previous)
	}
	return post
}

func (minds *Minds) GenerateSignPayload() (payload string) {
	payloadStruct := validator.H{
		"action":     string(minds.Action),
		"identity":   minds.Identity,
		"platform":   string(types.Platforms.Minds),
		"prev":       nil,
		"created_at": util.TimeToTimestampString(minds.CreatedAt),
		"uuid":       minds.Uuid.String(),
	}
	if minds.Previous != "" {
		payloadStruct["prev"] = minds.Previous
	}
	payloadBytes, err := json.Marshal(payloadStruct)
	if err != nil {
		l.Warnf("Error when marshaling struct: %s", err.Error())
		return ""
	}

	return string(payloadBytes)
}

func (minds *Minds) Validate() (err error) {
	// Minds username is case-insensitive
	minds.Identity = strings.ToLower(minds.Identity)
	minds.SignaturePayload = minds.GenerateSignPayload()
	post, err := minds.getContent()
	if err != nil {
		return err
	}
	minds.Text = post.Entities[0].Message

	return minds.validatePayload(post)
}

func (minds *Minds) getContent() (post *MindsPayload, err error) {
	url := fmt.Sprintf(URL, minds.ProofLocation)
	resp, err := http.Get(url)
	if err != nil {
		return nil, xerrors.Errorf("error when getting Minds post: %w", err)
	}
	if resp.StatusCode != 200 {
		return nil, xerrors.Errorf("error when requesting proof: Status code %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, xerrors.Errorf("Error when getting resp body: %w", err)
	}
	post = new(MindsPayload)
	err = json.Unmarshal(body, post)
	if err != nil {
		return nil, xerrors.Errorf("error when decoding JSON: %w", err)
	}
	if len(post.Entities) == 0 {
		return nil, xerrors.Errorf("Post not found")
	}
	return post, nil
}

func (minds *Minds) validatePayload(payload *MindsPayload) error {
	entity := payload.Entities[0]
	if minds.Identity != strings.ToLower(entity.Owner.UserName) {
		return xerrors.Errorf("Username mismatch: expect @%s, got @%s", minds.Identity, entity.Owner.UserName)
	}

	scanner := bufio.NewScanner(strings.NewReader(minds.Text))
	for scanner.Scan() {
		matched := re.FindStringSubmatch(scanner.Text())
		if len(matched) < 2 {
			continue // Search for next line
		}
		sigBase64 := matched[1]
		sigBytes, err := util.DecodeString(sigBase64)
		if err != nil {
			return xerrors.Errorf("Error when decoding signature %s: %s", sigBase64, err.Error())
		}
		minds.Signature = sigBytes
		return mycrypto.ValidatePersonalSignature(minds.SignaturePayload, sigBytes, minds.Pubkey)
	}

	return xerrors.Errorf("Signature not found in post text.")
}
