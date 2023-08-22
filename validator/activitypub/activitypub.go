package activitypub

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/nextdotid/proof_server/types"
	"github.com/nextdotid/proof_server/util"
	"github.com/nextdotid/proof_server/util/crypto"
	"github.com/nextdotid/proof_server/validator"
	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"
)

type ServerSoftware string

type NodeInfo struct {
	Links []NodeInfoLink `json:"links"`
}

type NodeInfoLink struct {
	Rel  string `json:"rel"`
	Href string `json:"href"`
}

type ActivityPub struct {
	*validator.Base
}

const (
	POST_TEMPLATE  = "Validate my ActivityPub identity @%s for Avatar 0x%s:\n\nSignature: %%SIG_BASE64%%\nUUID: %s\nPrevious: %s\nCreatedAt: %d\n\nPowered by Next.ID - Connect All Digital Identities.\n"
	MATCH_TEMPLATE = "^Signature: (.*)$"
)

var (
	l       = logrus.WithFields(logrus.Fields{"module": "validator", "validator": "activitypub"})
	re      = regexp.MustCompile(MATCH_TEMPLATE)
	Servers = struct {
		Mastodon ServerSoftware
		Misskey  ServerSoftware
		Pleroma  ServerSoftware
	}{
		Mastodon: "mastodon",
		Misskey:  "misskey",
		Pleroma:  "pleroma",
	}
)

func Init() {
	if validator.PlatformFactories == nil {
		validator.PlatformFactories = make(map[types.Platform]func(*validator.Base) validator.IValidator)
	}
	validator.PlatformFactories[types.Platforms.ActivityPub] = func(base *validator.Base) validator.IValidator {
		ap := ActivityPub{base}
		return &ap
	}
}

func (ap *ActivityPub) SplitID() (username, server string, err error) {
	// Trim initial @
	ap.Identity = strings.Trim(ap.Identity, "@")
	results := strings.Split(ap.Identity, "@")
	if len(results) != 2 {
		return "", "", xerrors.Errorf("invalid ActivityPub ID: Only one @ symbol should appear")
	}
	username = results[0]
	server = results[1]
	return username, server, nil
}

func (ap *ActivityPub) DetectServerSoftware() (server ServerSoftware, err error) {
	e := func(err error) error {
		return xerrors.Errorf("error when detecting server software: %w", err)
	}
	_, serverURL, err := ap.SplitID()
	if err != nil {
		return "", e(err)
	}
	// Get NodeInfo
	resp, err := http.Get(fmt.Sprintf("https://%s/.well-known/nodeinfo", serverURL))
	if err != nil {
		return "", e(err)
	}
	if resp.StatusCode != 200 {
		return "", xerrors.Errorf("error when detecting server software: node info returns %d", resp.StatusCode)
	}
	var nodeInfo NodeInfo
	err = json.NewDecoder(resp.Body).Decode(&nodeInfo)
	if err != nil {
		return "", e(err)
	}

	// Get true node info from links
	for _, link := range nodeInfo.Links {
		// TODO: maybe no need to be this strict?
		if link.Rel != "http://nodeinfo.diaspora.software/ns/schema/2.0" {
			continue
		}
		resp, err := http.Get(link.Href)
		if err != nil {
			return "", e(err)
		}
		if resp.StatusCode != 200 {
			return "", xerrors.Errorf("error when detecting server software: node info returns %d", resp.StatusCode)
		}
		var nodeInfo2 struct {
			Software struct {
				Name string `json:"name"`
			} `json:"software"`
		}
		err = json.NewDecoder(resp.Body).Decode(&nodeInfo2)
		if err != nil {
			return "", e(err)
		}
		switch nodeInfo2.Software.Name {
		case "mastodon":
			return Servers.Mastodon, nil
		case "misskey":
			return Servers.Misskey, nil
		case "pleroma":
			return Servers.Pleroma, nil
		default:
			return "", xerrors.Errorf("error when detecting server software: unsupported server: %s", nodeInfo2.Software.Name)
		}
	}
	return "", xerrors.Errorf("error when detecting server software: no supported node info link found")
}

func (ap *ActivityPub) GeneratePostPayload() (_ map[string]string) {
	id := ap.Identity
	previous := ap.Previous
	if previous == "" {
		previous = "null"
	}
	return map[string]string{
		"default": fmt.Sprintf(
			POST_TEMPLATE,
			id,
			crypto.CompressedPubkeyHex(ap.Pubkey),
			ap.Uuid.String(),
			previous,
			ap.CreatedAt.Unix(),
		),
	}
}

func (ap *ActivityPub) GenerateSignPayload() (payload string) {
	payloadStruct := validator.H{
		"action":     string(ap.Action),
		"identity":   ap.Identity,
		"platform":   string(types.Platforms.ActivityPub),
		"prev":       nil,
		"created_at": util.TimeToTimestampString(ap.CreatedAt),
		"uuid":       ap.Uuid.String(),
	}
	if ap.Previous != "" {
		payloadStruct["prev"] = ap.Previous
	}
	payloadBytes, err := json.Marshal(payloadStruct)
	if err != nil {
		l.Warnf("Error when marshaling struct: %s", err.Error())
		return ""
	}

	return string(payloadBytes)
}

func (ap *ActivityPub) Validate() (err error) {
	// Get Text
	server, err := ap.DetectServerSoftware()
	if err != nil {
		return err
	}
	switch server {
	case Servers.Mastodon:
	case Servers.Pleroma:
		err = ap.GetMastodonText()
	case Servers.Misskey:
		err = ap.GetMisskeyText()
	}
	if err != nil {
		return err
	}
	// Extract signature from text
	if err = ap.ExtractSignature(); err != nil {
		return err
	}
	// Verify signature
	return crypto.ValidatePersonalSignature(ap.GenerateSignPayload(), ap.Signature, ap.Pubkey)
}

func (ap *ActivityPub) GetAltID() (altID string) {
	return ap.AltID
}

func (ap *ActivityPub) ExtractSignature() (err error) {
	// Extract signature using regexp
	scanner := bufio.NewScanner(strings.NewReader(ap.Text))
	for scanner.Scan() {
		matches := re.FindStringSubmatch(scanner.Text())
		if len(matches) != 2 {
			continue // Search for next line
		}
		sig, err := base64.StdEncoding.DecodeString(matches[1])
		if err != nil {
			return xerrors.Errorf("error when parsing signature: %w", err)
		}
		ap.Signature = sig
		return nil
	}

	return xerrors.Errorf("no signature found")
}
