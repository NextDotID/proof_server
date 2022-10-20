package steam

import (
	"crypto/ecdsa"
	"io/ioutil"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/nextdotid/proof_server/types"
	"github.com/nextdotid/proof_server/validator"
	"github.com/samber/lo"
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

func generate() Steam {
	return Steam{
		&validator.Base{
			Platform:         types.Platforms.Steam,
			Previous:         "",
			Action:           types.Actions.Create,
			Pubkey:           &ecdsa.PublicKey{},
			Identity:         "BeFoRE-CS",
			AltID:            "",
			ProofLocation:    "",
			Signature:        []byte{},
			SignaturePayload: "",
			Text:             "",
			Extra:            map[string]string{},
			CreatedAt:        time.Now(), // TODO
			Uuid:             [16]byte{},
		},
	}
}

func Test_parseSteamXML(t *testing.T) {
	t.Run("error response", func(t *testing.T) {
		errResponse := "<?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"yes\"?><response><error><![CDATA[The specified profile could not be found.]]></error></response>"
		_, _, _, err := parseSteamXML([]byte(errResponse))
		require.Error(t, err)
		require.Contains(t, err.Error(), "The specified profile could not be found")
	})

	t.Run("correct response", func(t *testing.T) {
		response := getFileContent(t, FILENAME_NO_PAYLOAD)
		uid, username, description, err := parseSteamXML(response)
		require.NoError(t, err)
		require.Equal(t, "76561197968575517", uid)
		require.Equal(t, "ChetFaliszek", username)
		require.Contains(t, description, "<span")
	})
}

func Test_ExtractSteamID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		steamID := "76561198092541763"
		universe, userID, y, err := ExtractSteamID(steamID)
		require.NoError(t, err)
		require.Equal(t, uint(1), universe)
		require.Equal(t, uint(66138017), userID)
		require.Equal(t, uint(1), y)
	})

	t.Run("failure due to lack of universe", func(t *testing.T) {
		steamID := "4503604054613827" // validID - (1 << 56)
		_, _, _, err := ExtractSteamID(steamID)
		require.Error(t, err)
		require.Contains(t, err.Error(), "universe")
	})

	t.Run("failure due to wrong magic number", func(t *testing.T) {
		steamID := "72057594170203971" // validID - (0x100001 << 32)
		_, _, _, err := ExtractSteamID(steamID)
		require.Error(t, err)
		require.Contains(t, err.Error(), "account type")
	})
}

func Test_GetUserInfo(t *testing.T) {
	t.Run("success with CustomURL", func(t *testing.T) {
		steam := generate()
		steam.Identity = "BeFoRE-CS"
		require.NoError(t, steam.GetUserInfo())
		require.NotEqual(t, steam.Identity, steam.AltID)
		require.NotEqual(t, steam.Identity, "BeFoRE-CS")
	})

	t.Run("success with SteamID", func(t *testing.T) {
		steam := generate()
		steam.Identity = "76561198092541763"
		require.NoError(t, steam.GetUserInfo())
		require.NotEqual(t, steam.Identity, steam.AltID)
		require.NotEqual(t, steam.Identity, "BeFoRE-CS")
	})
}

func Test_GenearteSignPayload(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		steam := generate()
		steam.Identity = "BeFoRE-CS"
		payload := steam.GenerateSignPayload()
		require.Contains(t, payload, "76561198092541763") // real Identity
		require.NotContains(t, payload, "BeFoRE-CS")      // User-defined CustomURL should not appear
		require.Contains(t, payload, "null")              // Previous
	})
}

func Test_GeneratePostPayload(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		steam := generate()
		payload := steam.GeneratePostPayload()
		defaultPayload, ok := payload["default"]
		require.True(t, ok)
		lo.ForEach([]string{"NextID", steam.Uuid.String(), strconv.FormatInt(steam.CreatedAt.Unix(), 10)}, func(contains string, i int) {
			require.Contains(t, defaultPayload, contains)
		})
	})
}
