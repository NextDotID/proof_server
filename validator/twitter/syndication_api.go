package twitter

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type SyndicationAPIResponse struct {
	ID   string             `json:"id_str"`
	User SyndicationAPIUser `json:"user"`
	Text string             `json:"text"`
}

type SyndicationAPIUser struct {
	ID              string `json:"id_str"`
	Name            string `json:"name"`
	ProfileImageURL string `json:"profile_image_url"`
	ScreenName      string `json:"screen_name"`
	Verified        bool   `json:"verified"`
	IsBlueVerified  bool   `json:"is_blue_verified"`
}

const (
	TWEET_SINDICATION_API = "https://cdn.syndication.twimg.com/tweet-result?id=%s"
	USER_AGENT = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36"
)

func fetchPostWithSyndication(id string) (tweet *SyndicationAPIResponse, err error) {
	url := fmt.Sprintf(TWEET_SINDICATION_API, id)
	l.Debugf("URL: %s", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", USER_AGENT)

	resp, err := new(http.Client).Do(req)
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	tweet = new(SyndicationAPIResponse)
	err = json.Unmarshal(body, tweet)
	if err != nil {
		l.Warnf("Unmarshal Syndication body error: %s", string(body))
		return nil, err
	}

	return tweet, nil
}
