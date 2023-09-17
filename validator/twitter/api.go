package twitter

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	twitter "github.com/g8rswimmer/go-twitter/v2"
	"github.com/nextdotid/proof_server/config"
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
	twitterClient    *twitter.Client
	CurrentTokenList *TokenList
)

type authorize struct {
	Token string
}

func (a authorize) Add(req *http.Request) {
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", a.Token))
}

// Fetch tweet using twitter OAuth2.0 API.
// FIXME: should be switched to guest OAuth token solution.
func fetchPostWithAPI(id string, maxRetries int) (*APIResponse, error) {
	if twitterClient == nil {
		twitterClient = &twitter.Client{
			Authorizer: authorize{
				Token: config.C.Platform.Twitter.OauthToken,
			},
			Client: http.DefaultClient,
			Host:   "https://api.twitter.com",
		}
	}
	opts := twitter.TweetLookupOpts{
		Expansions:  []twitter.Expansion{twitter.ExpansionEntitiesMentionsUserName, twitter.ExpansionAuthorID},
		TweetFields: []twitter.TweetField{twitter.TweetFieldText, twitter.TweetFieldCreatedAt, twitter.TweetFieldEntities},
	}
	result, err := twitterClient.TweetLookup(context.Background(), []string{id}, opts)
	if err != nil {
		return nil, xerrors.Errorf("error when retriving tweet: %w", err)
	}
	tweet := result.Raw.Tweets[0]
	if tweet == nil {
		return nil, xerrors.Errorf("tweet not found: %s", id)
	}

	response := APIResponse{
		Text: tweet.Text,
	}
	response.User.ID = tweet.AuthorID
	if len(tweet.Entities.Mentions) > 0 {
		// Expect to be the user himself
		mention := tweet.Entities.Mentions[0]
		response.User.ScreenName = strings.ToLower(mention.UserName)
	}

	return &response, nil
}

// func fetchPostWithAPI(id string, maxRetries int) (tweet *APIResponse, err error) {
// 	const RETRY_AFTER = time.Second
// 	ctx := context.Background()
// 	if CurrentTokenList == nil {
// 		CurrentTokenList, err = GetTokenListFromS3(ctx)
// 		if err != nil {
// 			return nil, xerrors.Errorf("fetchPostWithAPI: %w", err)
// 		}
// 		if CurrentTokenList == nil {
// 			return nil, xerrors.Errorf("twitter token list does not exist")
// 		}
// 	}
// 	token := lo.Sample(CurrentTokenList.Tokens)
// 	if lo.IsEmpty(token.OAuthSecret) || lo.IsEmpty(token.OAuthKey) {
// 		return nil, xerrors.Errorf("twitter token seems to be empty")
// 	}

// 	// TODO: Use token to query specific tweet with twitter API
// 	// https://developer.twitter.com/en/docs/twitter-api/tweets/timelines/api-reference/get-users-id-tweets
// 	// https://api.twitter.com/1.1/statuses/show.json

// 	return nil, nil
// }
