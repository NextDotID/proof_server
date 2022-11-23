package headless_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/nextdotid/proof_server/headless"
)

func newValidRequest(location string, matchType string) headless.FindRequest {
	switch matchType {
	case "regexp":
		return headless.FindRequest{
			Location: location,
			Timeout:  "2s",
			Match: headless.Match{
				Type: "regexp",
				MatchRegExp: &headless.MatchRegExp{
					Selector: "*",
					Value:    "^Sig: .*$",
				},
			},
		}
	case "xpath":
		return headless.FindRequest{
			Location: location,
			Timeout:  "2s",
			Match: headless.Match{
				Type: "xpath",
				MatchXPath: &headless.MatchXPath{
					Selector: "//text()[contains(.,'Sig:')]",
				},
			},
		}
	case "js":
		return headless.FindRequest{
			Location: location,
			Timeout:  "2s",
			Match: headless.Match{
				Type: "js",
				MatchJS: &headless.MatchJS{
					Value: "() => [].filter.call(document.querySelectorAll('*'), (el) => el.textContent.includes('Sig:'))[0]",
				},
			},
		}
	}

	return headless.FindRequest{}
}

func Test_Find(t *testing.T) {
	apiTs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// To simulate network latency
		time.Sleep(time.Duration(300) * time.Millisecond)

		w.Header().Add("Content-Type", "application/json; charset=utf-8")
		w.Write([]byte(`{ "content": "Sig: match-this-text" }`))
	}))

	defer apiTs.Close()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		html := fmt.Sprintf(`
                <html>
                  <head>
                    <script>
                      fetch('%s')
                        .then(response => response.json())
                        .then(({ content }) => {
                            document.body.innerHTML = '<h1>' + content + '</h1>';
                        });
                    </script>
                  </head>
                  <body>
                  </body>
                </html>`,
			apiTs.URL)

		w.Header().Add("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(html))
	}))

	defer ts.Close()

	t.Run("success", func(t *testing.T) {
		// using regexp
		req := newValidRequest(ts.URL, "regexp")
		res := headless.FindRespond{}

		APITestCall(headless.Engine, "POST", "/v1/find", req, &res)

		assert.Equal(t, "Sig: match-this-text", res.Content)
		assert.Equal(t, "", res.Message)

		// using xpath
		req = newValidRequest(ts.URL, "xpath")
		res = headless.FindRespond{}

		APITestCall(headless.Engine, "POST", "/v1/find", req, &res)

		assert.Equal(t, "Sig: match-this-text", res.Content)
		assert.Equal(t, "", res.Message)

		// using js
		req = newValidRequest(ts.URL, "js")
		res = headless.FindRespond{}

		APITestCall(headless.Engine, "POST", "/v1/find", req, &res)

		assert.Equal(t, "Sig: match-this-text", res.Content)
		assert.Equal(t, "", res.Message)
	})

	t.Run("error ", func(t *testing.T) {
		// invalid location
		req := newValidRequest(ts.URL, "regexp")
		res := headless.FindRespond{}
		req.Location = ""
		APITestCall(headless.Engine, "POST", "/v1/find", req, &res)

		assert.Contains(t, res.Message, "location")

		// invalid timeout
		req = newValidRequest(ts.URL, "regexp")
		res = headless.FindRespond{}
		req.Timeout = "invalid"
		APITestCall(headless.Engine, "POST", "/v1/find", req, &res)

		assert.Contains(t, res.Message, "timeout")

		// invalid match type
		req = newValidRequest(ts.URL, "regexp")
		res = headless.FindRespond{}
		req.Match.Type = "invalid"
		APITestCall(headless.Engine, "POST", "/v1/find", req, &res)

		assert.Contains(t, res.Message, "match.type")

		// missing regexp value
		req = newValidRequest(ts.URL, "regexp")
		res = headless.FindRespond{}
		req.Match.MatchRegExp.Value = ""
		APITestCall(headless.Engine, "POST", "/v1/find", req, &res)

		assert.Contains(t, res.Message, "match.regexp.value")

		// missing xpath selector
		req = newValidRequest(ts.URL, "xpath")
		res = headless.FindRespond{}
		req.Match.MatchXPath.Selector = ""
		APITestCall(headless.Engine, "POST", "/v1/find", req, &res)

		assert.Contains(t, res.Message, "match.xpath.selector")

		// missing js value
		req = newValidRequest(ts.URL, "js")
		res = headless.FindRespond{}
		req.Match.MatchJS.Value = ""
		APITestCall(headless.Engine, "POST", "/v1/find", req, &res)

		assert.Contains(t, res.Message, "match.js.value")

		// target text is not found
		req = newValidRequest(ts.URL, "regexp")
		success := headless.FindRespond{}
		req.Match.MatchRegExp.Value = "unknown-text"
		APITestCall(headless.Engine, "POST", "/v1/find", req, &success)

		assert.Equal(t, success.Content, "")
	})
}
