package twitter

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_getAccessToken(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		require.NoError(t, getAccessToken())
		require.NotEmpty(t, accessToken)
		t.Logf("Access token: %s", accessToken)
	})
}

func Test_getGuestToken(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		require.NoError(t, getGuestToken())
		require.NotEmpty(t, guestToken)
		t.Logf("Guest token: %s", guestToken)
	})
}

func Test_FlowToken(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		require.NoError(t, getFlowToken())
		require.NotEmpty(t, flowToken)
		t.Logf("Flow token: %s", flowToken)
	})
}

func Test_OauthToken(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		require.NoError(t, GetOauthToken())
		t.Logf("Access token: %s", accessToken)
		t.Logf("Guest token: %s", guestToken)
		t.Logf("Flow token: %s", flowToken)
	})
}
