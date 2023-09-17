package twitter

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_getAccessToken(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tokens := new(Token)
		require.NoError(t, tokens.getAccessToken())
		require.NotEmpty(t, tokens.AccessToken)
		t.Logf("Access token: %s", tokens.AccessToken)
	})
}

func Test_getGuestToken(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tokens := new(Token)
		require.NoError(t, tokens.getGuestToken())
		require.NotEmpty(t, tokens.GuestToken)
		t.Logf("Guest token: %s", tokens.GuestToken)
	})
}

func Test_FlowToken(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tokens := new(Token)
		require.NoError(t, tokens.getFlowToken())
		require.NotEmpty(t, tokens.FlowToken)
		t.Logf("Flow token: %s", tokens.FlowToken)
	})
}

func Test_OauthToken(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tokens, err := GenerateOauthToken()
		require.NoError(t, err)
		require.False(t, tokens.IsExpired())
		t.Logf("Access token: %s", tokens.AccessToken)
		t.Logf("Guest token: %s", tokens.GuestToken)
		t.Logf("Flow token: %s", tokens.FlowToken)
	})
}
