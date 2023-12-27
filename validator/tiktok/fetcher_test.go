package tiktok

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_finalURLmatching(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		url := "https://www.tiktok.com/@scout_2015/video/6718335390845095173?_test=123"
		result := finalUrlRegexp.FindStringSubmatch(url)
		require.Equal(t, 3, len(result))
		require.Equal(t, "scout_2015", result[1])
		require.Equal(t, "6718335390845095173", result[2])
	})
}
