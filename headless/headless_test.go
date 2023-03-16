package headless_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nextdotid/proof_server/config"
	"github.com/nextdotid/proof_server/headless"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	config.Init("../config/config.test.json")
	headless.Init("", "")
	os.Exit(m.Run())
}

func APITestCall(engine *gin.Engine, method, url string, body any, response any) *httptest.ResponseRecorder {
	bb, _ := json.Marshal(body)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(method, url, bytes.NewReader(bb))
	req.Header.Add("Content-Type", "application/json")
	engine.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), response)

	return w
}

func Test_initUrlReplacementRule(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		flagReplacement := "abc.com=test.org,def.com=foobar.net"
		headless.InitUrlReplacementRule(flagReplacement)
		require.Equal(t, "test.org", headless.URLReplacement["abc.com"])
		require.Equal(t, "foobar.net", headless.URLReplacement["def.com"])
	})
}
