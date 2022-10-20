package steam

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"

	"github.com/nextdotid/proof_server/types"
	"github.com/nextdotid/proof_server/util"
	"github.com/nextdotid/proof_server/validator"
	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"
)

const (
	// Profile info by digit-based SteamID (e.g. "76561197968575517")
	PROFILE_PAGE_UID = "https://steamcommunity.com/profiles/%s/?xml=1"
	// Profile info by user-defined custom URL (e.g. "ChetFaliszek")
	PROFILE_PAGE_CUSTOM_URL = "https://steamcommunity.com/id/%s/?xml=1"
)

var (
	l = logrus.WithFields(logrus.Fields{"module": "validator", "validator": "steam"})
)

type Steam struct {
	*validator.Base
}

type SteamErrorResponse struct {
	XMLName xml.Name `xml:"response"`
	Error   string   `xml:"error"`
}

type SteamResponse struct {
	XMLName          xml.Name `xml:"profile"`
	SteamID64        string   `xml:"steamID64"`
	SteamID          string   `xml:"steamID"`
	StateMessage     string   `xml:"stateMessage"`
	PrivacyState     string   `xml:"privacyState"`
	VisibilityState  string   `xml:"visibilityState"`
	AvatarIcon       string   `xml:"avatarIcon"`
	AvatarMedium     string   `xml:"avatarMedium"`
	AvatarFull       string   `xml:"avatarFull"`
	VacBanned        string   `xml:"vacBanned"`
	TradeBanState    string   `xml:"tradeBanState"`
	IsLimitedAccount string   `xml:"isLimitedAccount"`
	CustomURL        string   `xml:"customURL"`
	MemberSince      string   `xml:"memberSince"`
	SteamRating      string   `xml:"steamRating"`
	HoursPlayed2Wk   string   `xml:"hoursPlayed2Wk"`
	Headline         string   `xml:"headline"`
	Location         string   `xml:"location"`
	Realname         string   `xml:"realname"`
	Summary          string   `xml:"summary"`
}

func Init() {
	if validator.PlatformFactories == nil {
		validator.PlatformFactories = make(map[types.Platform]func(*validator.Base) validator.IValidator)
	}
	validator.PlatformFactories[types.Platforms.Steam] = func(base *validator.Base) validator.IValidator {
		steam := Steam{base}
		return &steam
	}
}

func (steam *Steam) GeneratePostPayload() (post map[string]string) {
	return // TODO
}

func (steam *Steam) GenerateSignPayload() (payload string) {
	payloadStruct := validator.H{
		"action":     string(steam.Action),
		"identity":   steam.Identity,
		"platform":   string(types.Platforms.Steam),
		"prev":       nil,
		"created_at": util.TimeToTimestampString(steam.CreatedAt),
		"uuid":       steam.Uuid.String(),
	}
	if steam.Previous != "" {
		payloadStruct["prev"] = steam.Previous
	}
	payloadBytes, err := json.Marshal(payloadStruct)
	if err != nil {
		l.Warnf("Error when marshaling struct: %s", err.Error())
		return ""
	}

	return string(payloadBytes)

}

func (steam *Steam) Validate() (err error) {
	return xerrors.Errorf("TODO: Not implemented")
}

// GetUserInfo returns user info from steam profile page XML.
func (steam *Steam) GetUserInfo(isCustomUrl bool) (uid string, username string, description string, err error) {
	var url string
	if isCustomUrl {
		url = fmt.Sprintf(PROFILE_PAGE_CUSTOM_URL, steam.Identity)
	} else {
		url = fmt.Sprintf(PROFILE_PAGE_UID, steam.Identity)
	}

	resp, err := http.Get(url)
	if err != nil {
		return "", "", "", xerrors.Errorf("Error when getting steam profile page: %w", err)
	}
	if resp.StatusCode != 200 {
		return "", "", "", xerrors.Errorf("Error when getting steam profile page: status code %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", "", xerrors.Errorf("Error when reading steam profile page body: %w", err)
	}

	return parseSteamXML(body)
}

func parseSteamXML(xmlBody []byte) (uid string, username string, descripton string, err error) {
	errorResponse := new(SteamErrorResponse)
	err = xml.Unmarshal(xmlBody, errorResponse)
	if err == nil { // Error response
		return "", "", "", xerrors.Errorf("Error when fetching steam profile page: %s", errorResponse.Error)
	}

	response := new(SteamResponse)
	err = xml.Unmarshal(xmlBody, response)
	if err != nil {
		return "", "", "", xerrors.Errorf("Error when parsing steam profile page: %w", err)
	}

	return response.SteamID64, response.CustomURL, response.Summary, nil
}
