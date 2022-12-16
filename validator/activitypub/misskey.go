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
	Text string `json:"text"`
}

func (ap *ActivityPub)GetMisskeyText() (err error) {
	_, server, err := ap.SplitID()
	if err != nil {
		return err
	}

	body := misskeyNotesShowRequest {
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
	ap.Text = response.Text

	return nil // TODO
}
