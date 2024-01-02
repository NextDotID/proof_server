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

func Test_redirectToFinalURL(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		url := "https://www.tiktok.com/t/ZPRv3FPg5/"
		username, videoID, err := redirectToFinalURL(url, 0)
		require.NoError(t, err)
		require.Equal(t, "realwolfiesmom", username)
		require.Equal(t, "7287329983805197614", videoID)
	})
}

func Test_fetchOembedInfo(t *testing.T) {
	t.Run("full URL", func(t *testing.T) {
		url := "https://www.tiktok.com/@scout2015/video/6718335390845095173"
		result, err := fetchOembedInfo(url)
		require.NoError(t, err)
		require.Contains(t, result.Title, "Scramble up ur name")
	})

	t.Run("shortened URL", func(t *testing.T) {
		url := "https://www.tiktok.com/t/ZPRv3FPg5"
		result, err := fetchOembedInfo(url)
		require.NoError(t, err)
		require.Contains(t, result.EmbedProductID, "7287329983805197614")
	})
}
