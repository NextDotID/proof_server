package steam

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	FILENAME_NO_PAYLOAD = "./test_76561197968575517.xml"
)

func getFileContent(t *testing.T, filename string) []byte {
	file, err := os.Open(filename)
	require.NoError(t, err)
	defer file.Close()
	data, err := ioutil.ReadAll(file)
	require.NoError(t, err)
	return data
}

func Test_parseSteamXML(t *testing.T) {
	t.Run("error response", func(t *testing.T) {
		errResponse := "<?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"yes\"?><response><error><![CDATA[The specified profile could not be found.]]></error></response>"
		_, _, _, err := parseSteamXML([]byte(errResponse))
		require.Error(t, err)
		require.Contains(t, err.Error(), "The specified profile could not be found")
	})

	t.Run("correct response, but has no payload", func(t *testing.T) {
		response := getFileContent(t, FILENAME_NO_PAYLOAD)
		_, _, _, err := parseSteamXML(response)
		require.NoError(t, err)
	})
}
