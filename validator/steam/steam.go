package steam

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"

	"github.com/nextdotid/proof_server/types"
	"github.com/nextdotid/proof_server/util"
	"github.com/nextdotid/proof_server/util/crypto"
	"github.com/nextdotid/proof_server/validator"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"
)

const (
	// Profile info by digit-based SteamID (e.g. "76561197968575517")
	PROFILE_PAGE_STEAMID = "https://steamcommunity.com/profiles/%s/?xml=1"
	// Profile info by user-defined custom URL (e.g. "ChetFaliszek")
	PROFILE_PAGE_CUSTOM_URL = "https://steamcommunity.com/id/%s/?xml=1"
	// Ignore other part of the proof post, only take signature part.
	MATCH_TEMPLATE = "NextID proof: (.+?):"
)

var (
	l = logrus.WithFields(logrus.Fields{"module": "validator", "validator": "steam"})
	// Base64|CreatedAtTimestamp|UUID|Previous
	POST_STRUCT = map[string]string{
		"default": "NextID proof: %%SIG_BASE64%%:%d:%s:%s",
	}
	re = regexp.MustCompile(MATCH_TEMPLATE)
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
	post = make(map[string]string)
	previous := "null"
	if steam.Previous != "" {
		previous = steam.Previous
	}
	for langCode, template := range POST_STRUCT {
		post[langCode] = fmt.Sprintf(template, steam.CreatedAt.Unix(), steam.Uuid.String(), previous)
	}
	return post
}

func (steam *Steam) GenerateSignPayload() (payload string) {
	if err := steam.GetUserInfo(); err != nil {
		l.Errorf("Get user info failed: %s", err.Error())
		return ""
	}

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
	// steam.Text fetch included
	payload := steam.GenerateSignPayload()
	l.Debugf("Summary for user %s: %s", steam.Identity, steam.Text)
	if payload == "" {
		return xerrors.Errorf("error when generating sign payload")
	}
	found := re.FindAllStringSubmatch(steam.Text, 10) // Find up to 10 results
	if len(found) == 0 {
		return xerrors.Errorf("proof not found in user summary")
	}

	foundValid := false
	lo.ForEach(found, func(matched []string, index int) {
		if len(matched) != 2 {
			err = xerrors.Errorf("Invalid result on proof record No.%d", index + 1)
			return
		}
		sigBytes, matchedErr := util.DecodeString(matched[1])
		if matchedErr != nil {
			err = matchedErr
			return
		}

		validateErr := crypto.ValidatePersonalSignature(payload, sigBytes, steam.Pubkey)
		if validateErr != nil {
			err = validateErr
			return
		}

		foundValid = true
	})
	if foundValid {
		// At least 1 valid result is found. Ignore other errors.
		return nil
	}

	return
}

// See also: https://developer.valvesoftware.com/wiki/SteamID
func ExtractSteamID(idInDecimal string) (universe uint, userID uint, y uint, err error) {
	id, err := strconv.ParseUint(idInDecimal, 10, 64)
	if err != nil {
		return 0, 0, 0, xerrors.Errorf("parsing string SteamID: %w", err)
	}

	// Account type (4bit) <> Account instance (20bit)
	accountTypeAndInstance := (id >> 32) & 0x00FFFFFF
	// Account type 1 (individual)
	// Account instance 0x00001 (usually set to 1 for user accounts).
	if accountTypeAndInstance != uint64(0x00100001) {
		return 0, 0, 0, xerrors.Errorf("parsing SteamID: account type mismatch")
	}

	universe = uint((id >> 56) & 0xFF)    // First 8 bit
	userID = uint((id & 0xFFFFFFFE) >> 1) // [33, 63] bit
	y = uint(id & 0x1)                    // last 1 bit

	if universe != uint(1) { // Public
		return universe, userID, y, xerrors.Errorf("parsing SteamID: invalid universe identifier")
	}

	return
}

// GetUserInfo returns user info from steam profile page XML, will also refresh `self`'s `Identity`, `AltID` and `Text`.
func (steam *Steam) GetUserInfo() (err error) {
	if steam.Text != "" {
		// No duplicated fetching
		return nil
	}

	var url string
	_, _, _, steamIDErr := ExtractSteamID(steam.Identity)
	if steamIDErr != nil {
		l.Warnf("Error when parsing identity %s to steamID: %s", steam.Identity, steamIDErr.Error())
		url = fmt.Sprintf(PROFILE_PAGE_CUSTOM_URL, steam.Identity)
	} else {
		url = fmt.Sprintf(PROFILE_PAGE_STEAMID, steam.Identity)
	}

	resp, err := http.Get(url)
	if err != nil {
		return xerrors.Errorf("getting steam profile page: %w", err)
	}
	if resp.StatusCode != 200 {
		return xerrors.Errorf("getting steam profile page: status code %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return xerrors.Errorf("reading steam profile page body: %w", err)
	}

	uid, username, description, err := parseSteamXML(body)
	if err != nil {
		return err
	}

	steam.Identity = uid
	steam.AltID = username
	steam.Text = description
	return nil

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
