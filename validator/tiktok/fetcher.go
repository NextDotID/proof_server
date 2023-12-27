package tiktok

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"

	"golang.org/x/xerrors"
)

const (
	FINAL_URL_TEMPLATE = "^https://www\\.tiktok\\.com/@(.+?)/video/(\\d+)"
	OEMBED_URL_BASE    = "https://www.tiktok.com/oembed?url=https://www.tiktok.com/@%s/video/%s"
)

var (
	finalUrlRegexp = regexp.MustCompile(FINAL_URL_TEMPLATE)
)

type OEmbedInfo struct {
	// "version": "1.0",
	Version string `json:"version"`
	// "type": "video",
	Type string `json:"type"`
	// "title": "Scramble up ur name & I’ll try to guess it😍❤️ #foryoupage #petsoftiktok #aesthetic",
	Title string `json:"title"`
	// "author_url": "https://www.tiktok.com/@scout2015",
	AuthorURL string `json:"author_url"`
	// "author_name": "Scout & Suki",
	AuthorName string `json:"author_name"`
	// "width": "100%",
	Width string `json:"width"`
	// "height": "100%",
	Height string `json:"height"`
	// "html": "<blockquote
	Html string `json:"html"`
	// "thumbnail_width": 720,
	ThumbnailWidth int `json:"thumbnail_width"`
	// "thumbnail_height": 1280,
	ThumbnailHeight int `json:"thumbnail_height"`
	// "thumbnail_url": "https://p16.muscdn.com/obj/tos-maliva-p-0068/06kv6rfcesljdjr45ukb0000d844090v0200010605",
	ThumbnailUrl string `json:"thumbnail_url"`
	// "provider_url": "https://www.tiktok.com",
	ProviderUrl string `json:"provider_url"`
	// "provider_name": "TikTok"
	ProviderName string `json:"provider_name"`
}

// fetchOembedInfo fetches OEmbed card info from TikTok.
// Sample: `https://www.tiktok.com/oembed?url=https://www.tiktok.com/@scout2015/video/6718335390845095173`
func fetchOembedInfo(url string) (*OEmbedInfo, error) {
	username, videoID, err := redirectToFinalURL(url, 0)
	if err != nil {
		return nil, err
	}

	oembedURL := fmt.Sprintf(OEMBED_URL_BASE, username, videoID)
	resp, err := http.Get(oembedURL)
	if err != nil {
		return nil, xerrors.Errorf("tiktok: error when fetching oembed info: %w", err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, xerrors.Errorf("tiktok: error when reading oembed body: %w", err)
	}
	oembed := OEmbedInfo{}

	if err = json.Unmarshal(body, &oembed); err != nil {
		return nil, xerrors.Errorf("tiktok: error when parsing oembed body: %w", err)
	}

	return &oembed, nil
}

func redirectToFinalURL(url string, redirectCount int) (username, videoID string, err error) {
	l.WithField("count", redirectCount).Infof("Fetching: %s", url)
	const MAX_REDIRECT = 10
	if redirectCount > MAX_REDIRECT {
		return "", "", xerrors.Errorf("tiktok: too much redirect")
	}
	if username, videoID = parseFinalURL(url); username != "" {
		return username, videoID, nil
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", "", xerrors.Errorf("tiktok: HTTP error: %w", err)
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	resp, err := (&http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}).Do(req)
	if err != nil {
		return "", "", xerrors.Errorf("tiktok: HTTP error: %w", err)
	}

	redirectLocation, err := resp.Location()
	if redirectLocation != nil {
		return redirectToFinalURL(redirectLocation.String(), redirectCount+1)
	}
	return "", "", xerrors.Errorf("tiktok: not a valid URL")
}

func parseFinalURL(url string) (username, videoID string) {
	result := finalUrlRegexp.FindStringSubmatch(url)
	if len(result) != 3 {
		return "", ""
	}
	return result[1], result[2]
}
