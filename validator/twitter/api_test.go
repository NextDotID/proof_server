package twitter

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_fetchPostWithAPI(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tweet, err := fetchPostWithAPI("1652176440396517378", 10)
		require.NoError(t, err)
		require.Contains(t, tweet.Text, "Sig:")
		require.Equal(t, tweet.User.ScreenName, "bgm38")
		require.Equal(t, tweet.User.ID, "292254624")
	})
}
