package twitter

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/xerrors"
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

func fetchPostWithSyndication(id string, maxRetries int) (tweet *SyndicationAPIResponse, err error) {
	const RETRY_AFTER = time.Second

	url := fmt.Sprintf(TWEET_SINDICATION_API, id)
	l.Debugf("URL: %s", url)
	accumulatedErrors := ""
	for retry := 0; retry < maxRetries; retry++ {
		if retry != 0 {
			time.Sleep(RETRY_AFTER)
		}
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			accumulatedErrors += (err.Error() + "; ")
			continue
		}
		req.Header.Set("User-Agent", USER_AGENT)

		resp, err := new(http.Client).Do(req)
		if err != nil {
			accumulatedErrors += (err.Error() + "; ")
			continue
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			accumulatedErrors += (err.Error() + "; ")
			continue
		}
		tweet = new(SyndicationAPIResponse)
		err = json.Unmarshal(body, tweet)
		if err != nil {
			l.Warnf("Unmarshal Syndication body error: %s", string(body))
			accumulatedErrors += (err.Error() + ":" + string(body))
			continue
		}
		return tweet, nil
	}
	return nil, xerrors.Errorf("%d retries reached: %s", maxRetries, accumulatedErrors)
}
