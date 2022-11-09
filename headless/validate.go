package headless

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
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

type ValidateRequest struct {
	Location string `json:"location"`
	Timeout  string `json:"timeout"`
	Match    Match  `json:"match"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

type SuccessResponse struct {
	IsValid bool   `json:"is_valid"`
	Detail  string `json:"detail,omitempty"`
}

func errorResp(c *gin.Context, error_code int, err error) {
	c.JSON(error_code, ErrorResponse{
		Message: err.Error(),
	})
}

func validate(c *gin.Context) {
	var req ValidateRequest
	if err := c.Bind(&req); err != nil {
		errorResp(c, http.StatusBadRequest, xerrors.Errorf("Param error"))
		return
	}

	if err := checkValidateRequest(&req); err != nil {
		errorResp(c, http.StatusBadRequest, err)
		return
	}

	launcher := newLauncher(LauncherPath)
	defer launcher.Cleanup()
	defer launcher.Kill()

	u, err := launcher.Launch()
	if err != nil {
		errorResp(c, http.StatusInternalServerError, xerrors.Errorf("%w", err))
		return
	}

	browser := rod.New().ControlURL(u)
	if err := browser.Connect(); err != nil {
		errorResp(c, http.StatusInternalServerError, xerrors.Errorf("%w", err))
		return
	}

	defer browser.Close()

	page, err := browser.Page(proto.TargetCreateTarget{URL: req.Location})
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

	page = page.Timeout(timeoutDuration)
	if err := page.WaitLoad(); err != nil {
		errorResp(c, http.StatusInternalServerError, xerrors.Errorf("%w", err))
		return
	}

	if err := page.WaitRepaint(); err != nil {
		errorResp(c, http.StatusInternalServerError, xerrors.Errorf("%w", err))
		return
	}

	switch req.Match.Type {
	case matchTypeRegex:
		selector := req.Match.MatchRegExp.Selector
		if selector == "" {
			selector = "*"
		}

		if _, err := page.ElementR(selector, req.Match.MatchRegExp.Value); err != nil {
			c.JSON(http.StatusOK, SuccessResponse{IsValid: false, Detail: err.Error()})

			return
		}
	case matchTypeXPath:
		selector := req.Match.MatchXPath.Selector
		if _, err := page.ElementX(selector); err != nil {
			c.JSON(http.StatusOK, SuccessResponse{IsValid: false, Detail: err.Error()})

			return
		}
	case matchTypeJS:
		js := req.Match.MatchJS.Value
		if _, err := page.ElementByJS(rod.Eval(js)); err != nil {
			c.JSON(http.StatusOK, SuccessResponse{IsValid: false, Detail: err.Error()})

			return
		}
	}

	c.JSON(http.StatusOK, SuccessResponse{IsValid: true})
}

func checkValidateRequest(req *ValidateRequest) error {
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
