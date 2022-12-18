package activitypub

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/xerrors"
)

type misskeyNotesShowRequest struct {
	NoteID string `json:"noteId"`
}

// Only focus on `text` field for now.
// TODO: show error message if it is not public.
type misskeyNotesShowResponse struct {
	User misskeyNotesShowResponseUser `json:"user"`
	Text string                       `json:"text"`
}

type misskeyNotesShowResponseUser struct {
	Id       string `json:"id"`
	Username string `json:"username"`
}

func (ap *ActivityPub) GetMisskeyText() (err error) {
	_, server, err := ap.SplitID()
	if err != nil {
		return err
	}

	body := misskeyNotesShowRequest{
		NoteID: ap.ProofLocation,
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}
	resp, err := http.Post(fmt.Sprintf("https://%s/api/notes/show", server), "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		return xerrors.Errorf("error when fetching Misskey note: %w", err)
	}
	var response misskeyNotesShowResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return xerrors.Errorf("error when decoding Misskey note response: %w", err)
	}
	postIdentity := fmt.Sprintf("%s@%s", response.User.Username, server)
	if postIdentity != ap.Identity {
		return xerrors.Errorf("Error when fetching Misskey note: This post is made by %s, not %s", postIdentity, ap.Identity)
	}

	ap.AltID = response.User.Id
	ap.Text = response.Text
	return nil
}
