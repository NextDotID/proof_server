package twitter

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_SyndicationAPI(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// before_each(t)
		tweet, err := fetchPostWithSyndication("1652176440396517378")
		require.Nil(t, err)
		require.Contains(t, tweet.Text, "Sig:")
		require.Equal(t, "292254624", tweet.User.ID)
	})
}
