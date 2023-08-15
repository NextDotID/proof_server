package twitter

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_SyndicationAPI(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Read /tmp/result_without_timeline.json file and deserialize it into GraphQLResponse
		// Then compare it with expected GraphQLResponse
		postID := "1687007065032814593"
		result, err := fetchPostWithSyndication(postID, 1)
		require.NoError(t, err)
		require.Equal(t, postID, result.ID)
		require.Equal(t, "bgm38", result.User.ScreenName)
		require.Equal(t, "292254624", result.User.ID)
	})
}
