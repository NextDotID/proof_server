package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nextdotid/proof-server/config"
	"github.com/nextdotid/proof-server/model"
	"github.com/nextdotid/proof-server/validator/ethereum"
	"github.com/nextdotid/proof-server/validator/twitter"
	"github.com/nextdotid/proof-server/validator/keybase"
	"github.com/nextdotid/proof-server/validator/github"
)

func before_each(t *testing.T)  {
	// Clean DB
	model.DB.Where("1 = 1").Delete(&model.Proof{})
	model.DB.Where("1 = 1").Delete(&model.ProofChain{})
}

func TestMain(m *testing.M)  {
	config.Init("../config/config.test.json")
	model.Init()
	Init()

	twitter.Init()
	keybase.Init()
	ethereum.Init()
	github.Init()

	before_each(nil)

	os.Exit(m.Run())
}

func APITestCall(engine *gin.Engine, method, url string, body interface{}, response interface{}) *httptest.ResponseRecorder {
	body_bytes, _ := json.Marshal(body)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(method, url, bytes.NewReader(body_bytes))
	engine.ServeHTTP(w, req)

	json.Unmarshal(w.Body.Bytes(), response)
	return w
}
