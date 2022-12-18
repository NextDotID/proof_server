package activitypub

import (
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/xerrors"
)

const MASTODON_API_STATUS = "https://%s/api/v1/statuses/%s"

type MastodonResponse struct {
	Account MastodonResponseAccount `json:"account"`
	Content string `json:"content"`
}

type MastodonResponseAccount struct {
	// ASCII
	Username string `json:"username"`
	// Digits
	Id string `json:"id"`
}

// GetMastodonText can also deal with Pleroma server.
func (ap *ActivityPub) GetMastodonText() (err error) {
	_, server, err := ap.SplitID()
	if err != nil {
		return err
	}
	resp, err := http.Get(fmt.Sprintf(MASTODON_API_STATUS, server, ap.ProofLocation))
	if err != nil {
		return xerrors.Errorf("failed to get mastodon / pleroma status: %w", err)
	}
	var response MastodonResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return xerrors.Errorf("failed to decode mastodon / pleroma status: %w", err)
	}

	postIdentity := fmt.Sprintf("@%s@%s", response.Account.Username, server)
	if postIdentity != ap.Identity {
		return xerrors.Errorf("failed to identify mastodon / pleroma status: identity mismatch: %s != %s", postIdentity, ap.Identity)
	}

	ap.AltID = response.Account.Id
	ap.Text = response.Content
	return nil
}
