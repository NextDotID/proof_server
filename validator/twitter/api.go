package twitter

import (
	"context"
	"time"

	"github.com/samber/lo"
	"golang.org/x/xerrors"
)

type APIResponse struct {
	User struct {
		ID         string `json:"user_id"`
		ScreenName string `json:"screen_name"`
	} `json:"user"`
	Text string `json:"text"`
}

var (
	CurrentTokenList *TokenList
)

func fetchPostWithAPI(id string, maxRetries int) (tweet *APIResponse, err error) {
	const RETRY_AFTER = time.Second
	ctx := context.Background()
	if CurrentTokenList == nil {
		CurrentTokenList, err = GetTokenListFromS3(ctx)
		if err != nil {
			return nil, xerrors.Errorf("fetchPostWithAPI: %w", err)
		}
		if CurrentTokenList == nil {
			return nil, xerrors.Errorf("twitter token list does not exist")
		}
	}
	token := lo.Sample(CurrentTokenList.Tokens)
	if lo.IsEmpty(token.OAuthSecret) || lo.IsEmpty(token.OAuthKey) {
		return nil, xerrors.Errorf("twitter token seems to be empty")
	}

	// TODO: Use token to query specific tweet with twitter API
	// https://developer.twitter.com/en/docs/twitter-api/tweets/timelines/api-reference/get-users-id-tweets
	// https://api.twitter.com/1.1/statuses/show.json

	return nil, nil
}
