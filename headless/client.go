package headless

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"path"

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

// Find find whether the target matching payload exists
func (h *HeadlessClient) Find(ctx context.Context, payload *FindRequest) (string, error) {
	u, err := url.Parse(h.url)
	if err != nil {
		return "", xerrors.Errorf("%w", err)
	}

	u.Path = path.Join(u.Path, "/v1/find")
	body, err := json.Marshal(payload)
	if err != nil {
		return "", xerrors.Errorf("%w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewReader(body))
	if err != nil {
		return "", xerrors.Errorf("%w", err)
	}

	req.Header.Add("Content-Type", "application/json")

	res, err := h.client.Do(req)
	if res != nil && err != nil {
		if _, err := io.Copy(io.Discard, res.Body); err != nil {
			return "", xerrors.Errorf("%w", err)
		}
	}

	if res != nil {
		defer res.Body.Close()
	}

	if err != nil {
		return "", xerrors.Errorf("%w", err)
	}

	contents, err := io.ReadAll(res.Body)
	if err != nil {
		return "", xerrors.Errorf("%w", err)
	}

	var resBody FindRespond
	if err := json.Unmarshal(contents, &resBody); err != nil {
		return "", xerrors.Errorf("%w", err)
	}

	return resBody.Content, nil
}
