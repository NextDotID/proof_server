package headless

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"golang.org/x/xerrors"
)

// HeadlessClient handles communication for headless browser service
type HeadlessClient struct {
	url    string
	client *http.Client
}

// NewHeadlessClient creates a new headless client
func NewHeadlessClient(url string) *HeadlessClient {
	return &HeadlessClient{url, http.DefaultClient}
}

// Validate validates whether the given payload is valid
func (h *HeadlessClient) Validate(ctx context.Context, payload *ValidateRequest) (bool, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return false, xerrors.Errorf("%w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.url, bytes.NewReader(body))
	if err != nil {
		return false, xerrors.Errorf("%w", err)
	}

	req.Header.Add("Content-Type", "application/json")

	res, err := h.client.Do(req)
	if res != nil && err != nil {
		if _, err := io.Copy(io.Discard, res.Body); err != nil {
			return false, xerrors.Errorf("%w", err)
		}
	}

	if res != nil {
		defer res.Body.Close()
	}

	if err != nil {
		return false, xerrors.Errorf("%w", err)
	}

	contents, err := io.ReadAll(res.Body)
	if err != nil {
		return false, xerrors.Errorf("%w", err)
	}

	var resBody ValidateRespond
	if err := json.Unmarshal(contents, &resBody); err != nil {
		return false, xerrors.Errorf("%w", err)
	}

	return resBody.IsValid, nil
}
