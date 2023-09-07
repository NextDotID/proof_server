package twitter

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/samber/lo"
	"golang.org/x/xerrors"
)

type SyndicationAPIResponse struct {
	ID   string             `json:"id_str"`
	User SyndicationAPIUser `json:"user"`
	Text string             `json:"text"`
}

type SyndicationAPIUser struct {
	ID         string `json:"id_str"`
	ScreenName string `json:"screen_name"`
}

// / data.threaded_conversation_with_injections.instructions[0].entries[?].content.itemContent.tweet_results.result.legacy.full_text
// / data.threaded_conversation_with_injections.instructions[0].entries[?].content.itemContent.tweet_results.result.core.user_results.result.legacy.name
// / ?: find entryId = "tweet-TWEETID"
type GraphQLAPIResponse struct {
	Data struct {
		ThreadedConversationWithInjections struct {
			Instructions []struct {
				Entries []GraphQLAPIEntry `json:"entries"`
			} `json:"instructions"`
		} `json:"threaded_conversation_with_injections"`
	} `json:"data"`
}

type GraphQLAPIEntry struct {
	EntryID string `json:"entryId"`
	Content struct {
		ItemContent struct {
			TweetResults struct {
				Result struct {
					Core struct {
						UserResults struct {
							Result struct {
								RestID string `json:"rest_id"`
								Legacy struct {
									ScreenName string `json:"screen_name"`
								} `json:"legacy"`
							} `json:"result"`
						} `json:"user_results"`
					} `json:"core"`
					Legacy struct {
						FullText string `json:"full_text"`
					} `json:"legacy"`
				} `json:"result"`
			} `json:"tweet_results"`
		} `json:"itemContent"`
	} `json:"content"`
}

const (
	GUEST_TOKEN_REQUEST = "Bearer AAAAAAAAAAAAAAAAAAAAAPYXBAAAAAAACLXUNDekMxqa8h%2F40K4moUkGsoc%3DTYfbDKbT3jJPCEVnMYqilB28NHfOPqkca3qaAxGfsyKCs0wRbw"

	QUERY_URL_HEAD = "https://api.twitter.com/graphql/miKSMGb2R1SewIJv2-ablQ/TweetDetail?variables=%7B%22focalTweetId%22%3A%22"
	QUERY_URL_TAIL = "%22,%22withBirdwatchNotes%22%3Afalse,%22includePromotedContent%22%3Afalse,%22withDownvotePerspective%22%3Afalse,%22withReactionsMetadata%22%3Afalse,%22withReactionsPerspective%22%3Afalse,%22withVoice%22%3Afalse,%22withV2Timeline%22%3Afalse%7D&features=%7B%22blue_business_profile_image_shape_enabled%22%3Afalse,%22rweb_lists_timeline_redesign_enabled%22%3Atrue,%22responsive_web_graphql_exclude_directive_enabled%22%3Atrue,%22verified_phone_label_enabled%22%3Afalse,%22creator_subscriptions_tweet_preview_api_enabled%22%3Atrue,%22responsive_web_graphql_timeline_navigation_enabled%22%3Afalse,%22responsive_web_graphql_skip_user_profile_image_extensions_enabled%22%3Afalse,%22tweetypie_unmention_optimization_enabled%22%3Afalse,%22vibe_api_enabled%22%3Afalse,%22responsive_web_edit_tweet_api_enabled%22%3Afalse,%22graphql_is_translatable_rweb_tweet_is_translatable_enabled%22%3Afalse,%22view_counts_everywhere_api_enabled%22%3Afalse,%22longform_notetweets_consumption_enabled%22%3Atrue,%22tweet_awards_web_tipping_enabled%22%3Afalse,%22freedom_of_speech_not_reach_fetch_enabled%22%3Afalse,%22standardized_nudges_misinfo%22%3Afalse,%22tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled%22%3Afalse,%22interactive_text_enabled%22%3Afalse,%22responsive_web_text_conversations_enabled%22%3Afalse,%22longform_notetweets_rich_text_read_enabled%22%3Afalse,%22longform_notetweets_inline_media_enabled%22%3Afalse,%22responsive_web_enhance_cards_enabled%22%3Afalse%7D"
	USER_AGENT     = "User-Agent: Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)"
)

// something like "1691388211879432192"
var GuestToken string

func fetchPostWithSyndication(id string, maxRetries int) (tweet *SyndicationAPIResponse, err error) {
	const RETRY_AFTER = time.Second

	accumulatedErrors := ""
	for retry := 0; retry < maxRetries; retry++ {
		if retry != 0 {
			time.Sleep(RETRY_AFTER)
		}
		// Fetching guestToken
		err := fetchGuestToken()
		if err != nil {
			accumulatedErrors += (err.Error() + "; ")
			continue
		}

		tweet, err := fetchPost(id)
		if err != nil {
			accumulatedErrors += (err.Error() + "; ")
			continue
		}
		return tweet, nil
	}
	return nil, xerrors.Errorf("%d retries reached: %s", maxRetries, accumulatedErrors)
}

func fetchGuestToken() (err error) {
	if GuestToken != "" {
		return nil
	}
	req, err := http.NewRequest("POST", "https://api.twitter.com/1.1/guest/activate.json", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", GUEST_TOKEN_REQUEST)
	req.Header.Set("User-Agent", USER_AGENT)

	resp, err := new(http.Client).Do(req)
	if err != nil {
		return err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var guestTokenResponse struct {
		GuestToken string `json:"guest_token"`
	}
	err = json.Unmarshal(body, &guestTokenResponse)
	if err != nil {
		return err
	}
	if guestTokenResponse.GuestToken == "" {
		return xerrors.Errorf("Guest token is empty")
	}

	GuestToken = guestTokenResponse.GuestToken

	return nil
}

func fetchPost(postID string) (post *SyndicationAPIResponse, err error) {
	req, err := http.NewRequest("GET", QUERY_URL_HEAD+postID+QUERY_URL_TAIL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", GUEST_TOKEN_REQUEST)
	req.Header.Set("User-Agent", USER_AGENT)
	req.Header.Set("x-guest-token", GuestToken)

	resp, err := new(http.Client).Do(req)
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	response := new(GraphQLAPIResponse)
	err = json.Unmarshal(body, response)
	if err != nil {
		return nil, err
	}
	if len(response.Data.ThreadedConversationWithInjections.Instructions) == 0 {
		return nil, xerrors.Errorf("No instructions found in response")
	}
	instruction := response.Data.ThreadedConversationWithInjections.Instructions[0]
	entry, found := lo.Find(instruction.Entries, func(entry GraphQLAPIEntry) bool {
		return entry.EntryID == ("tweet-" + postID)
	})
	if !found {
		return nil, xerrors.Errorf("Tweet specified in ProofLocation is not found in API response")
	}

	return &SyndicationAPIResponse{
		ID: postID,
		User: SyndicationAPIUser{
			ID:         entry.Content.ItemContent.TweetResults.Result.Core.UserResults.Result.RestID,
			ScreenName: entry.Content.ItemContent.TweetResults.Result.Core.UserResults.Result.Legacy.ScreenName,
		},
		Text: entry.Content.ItemContent.TweetResults.Result.Legacy.FullText,
	}, nil
}
