package headless

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/nextdotid/proof_server/common"
	"github.com/ssoroka/slice"
	"golang.org/x/xerrors"
)

const (
	matchTypeRegex = "regexp"
	matchTypeXPath = "xpath"
	matchTypeJS    = "js"
	defaultTimeout = "10s"
)

var (
	validMatchTypes = map[string]struct{}{
		matchTypeJS:    {},
		matchTypeXPath: {},
		matchTypeRegex: {},
	}
)

type MatchRegExp struct {
	// Selector is the target element if not specified "*" will be used
	Selector string `json:"selector"`

	// Value is the target value
	Value string `json:"value"`
}

type MatchXPath struct {
	// Selector is the xpath selector
	Selector string `json:"selector"`
}

type MatchJS struct {
	// Value is the javascript value
	Value string `json:"value"`
}

type Match struct {
	Type        string       `json:"type"`
	MatchRegExp *MatchRegExp `json:"regexp"`
	MatchXPath  *MatchXPath  `json:"xpath"`
	MatchJS     *MatchJS     `json:"js"`
}

type FindRequest struct {
	Location string `json:"location"`
	Timeout  string `json:"timeout"`
	Match    Match  `json:"match"`
}

type FindRespond struct {
	Content string `json:"content"`
	Message string `json:"message,omitempty"`
}

func errorResp(c *gin.Context, error_code int, err error) {
	c.JSON(error_code, FindRespond{
		Message: err.Error(),
	})
}

func validate(c *gin.Context) {
	var req FindRequest
	if err := c.Bind(&req); err != nil {
		errorResp(c, http.StatusBadRequest, xerrors.Errorf("Param error: %v", err))
		return
	}

	if err := checkValidateRequest(&req); err != nil {
		errorResp(c, http.StatusBadRequest, err)
		return
	}

	var launcher *launcher.Launcher
	switch common.CurrentRuntime {
	case common.Runtimes.Lambda:
		launcher = newLambdaLauncher(LauncherPath)
	case common.Runtimes.Standalone:
		launcher = newLauncher(LauncherPath)
	}

	defer launcher.Kill()
	defer launcher.Cleanup()

	u, err := launcher.Launch()
	if err != nil {
		errorResp(c, http.StatusInternalServerError, xerrors.Errorf("%w", err))
		return
	}

	browser := rod.New().ControlURL(u)
	defer browser.Close()
	if err := browser.Connect(); err != nil {
		errorResp(c, http.StatusInternalServerError, xerrors.Errorf("%w", err))
		return
	}

	page, err := browser.Page(proto.TargetCreateTarget{URL: ReplaceLocation(req.Location)})
	if err != nil {
		errorResp(c, http.StatusInternalServerError, xerrors.Errorf("%w", err))
		return
	}

	timeout := req.Timeout
	if timeout == "" {
		timeout = defaultTimeout
	}

	timeoutDuration, err := time.ParseDuration(timeout)
	if err != nil {
		errorResp(c, http.StatusInternalServerError, xerrors.Errorf("%w", err))
		return
	}
	router := page.HijackRequests()

	resources := []proto.NetworkResourceType{
		proto.NetworkResourceTypeFont,
		proto.NetworkResourceTypeImage,
		proto.NetworkResourceTypeMedia,
		proto.NetworkResourceTypeStylesheet,
		proto.NetworkResourceTypeWebSocket, // we don't need websockets to fetch html
	}

	router.MustAdd("*", func(ctx *rod.Hijack) {
		if slice.Contains(resources, ctx.Request.Type()) {
			ctx.Response.Fail(proto.NetworkErrorReasonBlockedByClient)
			return
		}

		ctx.ContinueRequest(&proto.FetchContinueRequest{})
	})

	go router.Run()
	defer router.Stop()

	page = page.Timeout(timeoutDuration)
	if err := page.WaitLoad(); err != nil {
		errorResp(c, http.StatusInternalServerError, xerrors.Errorf("%w", err))
		return
	}

	// Wait for XHR
	page.WaitNavigation(proto.PageLifecycleEventNameNetworkAlmostIdle)()
	content, err := find(req.Match, page)
	if err != nil {
		c.JSON(http.StatusOK, FindRespond{Content: "", Message: err.Error()})

		return
	}

	c.JSON(http.StatusOK, FindRespond{Content: content})
}

func find(match Match, page *rod.Page) (content string, err error) {
	var element *rod.Element
	switch match.Type {
	case matchTypeRegex:
		if match.MatchRegExp.Selector == "" {
			match.MatchRegExp.Selector = "*"
		}

		element, err = page.ElementR(match.MatchRegExp.Selector, match.MatchRegExp.Value)
		if err != nil {
			return "", xerrors.Errorf("%w", err)
		}
	case matchTypeXPath:
		element, err = page.ElementX(match.MatchXPath.Selector)
		if err != nil {
			return "", xerrors.Errorf("%w", err)
		}
	case matchTypeJS:
		element, err = page.ElementByJS(rod.Eval(match.MatchJS.Value))
		if err != nil {
			return "", xerrors.Errorf("%w", err)
		}
	default:
		return "", xerrors.Errorf("%s", "invalid payload")
	}

	text, err := element.Text()
	if err != nil {
		return "", xerrors.Errorf("%w", err)
	}

	return text, nil
}

func checkValidateRequest(req *FindRequest) error {
	if req.Location == "" {
		return xerrors.Errorf("'location' is missing")
	}

	if req.Timeout != "" {
		if _, err := time.ParseDuration(req.Timeout); err != nil {
			return xerrors.Errorf("'timeout' is invalid")
		}
	}

	if _, ok := validMatchTypes[req.Match.Type]; !ok {
		return xerrors.Errorf("'match.type' should be 'regexp', 'xpath', or 'js'")
	}

	if req.Match.Type == matchTypeRegex {
		if req.Match.MatchRegExp == nil {
			return xerrors.Errorf("'match.regexp' payload is missing")
		}

		if req.Match.MatchRegExp.Value == "" {
			return xerrors.Errorf("'match.regexp.value' must be specified")
		}
	}

	if req.Match.Type == matchTypeXPath {
		if req.Match.MatchXPath == nil {
			return xerrors.Errorf("'match.xpath' payload is missing")
		}

		if req.Match.MatchXPath.Selector == "" {
			return xerrors.Errorf("'match.xpath.selector' must be specified")
		}
	}

	if req.Match.Type == matchTypeJS {
		if req.Match.MatchJS == nil {
			return xerrors.Errorf("'match.js' payload is missing")
		}

		if req.Match.MatchJS.Value == "" {
			return xerrors.Errorf("'match.js.value' must be specified")
		}
	}

	return nil
}

/// ReplaceLocation uses URLReplacement as rule to replace part of the original URL.
func ReplaceLocation(originalURL string) string {
	if len(URLReplacement) == 0 {
		return originalURL
	}

	for original, replacement := range URLReplacement {
		originalURL = strings.ReplaceAll(originalURL, original, replacement)
	}

	return originalURL
}
