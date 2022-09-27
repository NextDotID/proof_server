package dns

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_query(t *testing.T) {
	_, err := query("nonexist.example.com")
	require.Error(t, err)

	body, err := query("example.com")
	require.NoError(t, err)
	require.NotNil(t, body)
}
