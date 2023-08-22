package dns

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nextdotid/proof_server/types"
	"github.com/nextdotid/proof_server/util"
	"github.com/nextdotid/proof_server/util/crypto"
	"github.com/nextdotid/proof_server/validator"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"
)

const (
	DOH = "https://cloudflare-dns.com/dns-query?type=TXT&name=%s"
)

// https://developers.cloudflare.com/1.1.1.1/encryption/dns-over-https/make-api-requests/dns-json/
type DOHResponse struct {
	// The Response Code of the DNS Query. These are defined here: https://www.iana.org/assignments/dns-parameters/dns-parameters.xhtml#dns-parameters-6Open external link.
	Status uint `json:"Status"`
	//If true, it means the truncated bit was set. This happens when the DNS answer is larger than a single UDP or TCP packet. TC will almost always be false with Cloudflare DNS over HTTPS because Cloudflare supports the maximum response size.
	TC bool `json:"TC"`
	// If true, it means the Recursive Desired bit was set. This is always set to true for Cloudflare DNS over HTTPS.
	RD bool `json:"RD"`
	// If true, it means the Recursion Available bit was set. This is always set to true for Cloudflare DNS over HTTPS.
	RA bool `json:"RA"`
	// If true, it means that every record in the answer was verified with DNSSEC.
	AD bool `json:"AD"`
	// If true, the client asked to disable DNSSEC validation. In this case, Cloudflare will still fetch the DNSSEC-related records, but it will not attempt to validate the records.
	CD       bool          `json:"CD"`
	Question []DOHQuestion `json:"Question"`
	// If Answer is empty, this field will appear.
	Authority *[]DOHAnswer `json:"Authority"`
	Answer    *[]DOHAnswer `json:"Answer"`
}

type DOHQuestion struct {
	// The record name requested.
	Name string `json:"name"`
	// The type of DNS record requested. These are defined here: https://www.iana.org/assignments/dns-parameters/dns-parameters.xhtml#dns-parameters-4Open external link.
	Type int `json:"type"`
}

type DOHAnswer struct {
	// The record owner.
	Name string `json:"name"`
	// The type of DNS record. These are defined here: https://www.iana.org/assignments/dns-parameters/dns-parameters.xhtml#dns-parameters-4Open external link.
	Type int `json:"type"`
	// The number of seconds the answer can be stored in cache before it is considered stale.
	TTL uint `json:"ttl"`
	// The value of the DNS record for the given name and type. The data will be in text for standardized record types and in hex for unknown types.
	Data string `json:"data"`
}

type TXTPayload struct {
	Version   uint
	Signature string
	CreatedAt time.Time
	uuid      uuid.UUID
	Previous  *string
}

type DNS struct {
	*validator.Base
}

const (
	TXT_PAYLOAD_V1 = "ps:true;v:1;sig:%s;ca:%d;uuid:%s;prev:%s"
)

var (
	l = logrus.WithFields(logrus.Fields{"module": "validator", "validator": "dns"})
)

func Init() {
	if validator.PlatformFactories == nil {
		validator.PlatformFactories = make(map[types.Platform]func(*validator.Base) validator.IValidator)
	}
	validator.PlatformFactories[types.Platforms.DNS] = func(base *validator.Base) validator.IValidator {
		dns := DNS{base}
		return &dns
	}

}

func (dns *DNS) GeneratePostPayload() (post map[string]string) {
	var previous string
	if dns.Previous != "" {
		previous = dns.Previous
	} else {
		previous = "null"
	}
	return map[string]string{
		"default": fmt.Sprintf(TXT_PAYLOAD_V1, "%SIG_BASE64%", dns.CreatedAt.Unix(), dns.Uuid.String(), previous),
	}
}

func (dns *DNS) GenerateSignPayload() (payload string) {
	payloadStruct := validator.H{
		"action":     string(dns.Action),
		"identity":   strings.ToLower(dns.Identity),
		"platform":   string(types.Platforms.DNS),
		"prev":       nil,
		"created_at": util.TimeToTimestampString(dns.CreatedAt),
		"uuid":       dns.Uuid.String(),
	}
	if dns.Previous != "" {
		payloadStruct["prev"] = dns.Previous
	}
	payloadBytes, err := json.Marshal(payloadStruct)
	if err != nil {
		l.Warnf("Error when marshaling struct: %s", err.Error())
		return ""
	}

	return string(payloadBytes)
}

func (dns *DNS) Validate() (err error) {
	// domain name is case-insensitive
	dns.Identity = strings.ToLower(dns.Identity)
	dns.AltID = dns.Identity
	dns.SignaturePayload = dns.GenerateSignPayload()
	query_resp, err := query(dns.Identity)
	if err != nil {
		return err
	}
	txt, found := lo.Find(*query_resp.Answer, func(txt DOHAnswer) bool {
		result, parse_err := parseTxt(txt.Data)
		return parse_err == nil && result.uuid == dns.Uuid
	})
	if !found {
		return xerrors.New("matched TXT record not found.")
	}
	payload, _ := parseTxt(txt.Data)
	dns.Text = txt.Data
	dns.Signature, err = base64.StdEncoding.DecodeString(payload.Signature)
	if err != nil {
		return xerrors.New("sig in TXT record cannot be recognized.")
	}

	return crypto.ValidatePersonalSignature(dns.SignaturePayload, dns.Signature, dns.Pubkey)
}

func (dns *DNS) GetAltID() string {
	return dns.AltID
}

func query(domain string) (doh_response *DOHResponse, err error) {
	req, err := http.NewRequest("GET", fmt.Sprintf(DOH, domain), nil)
	req.Header.Set("Accept", "application/dns-json")
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, xerrors.Errorf("status code %d", resp.StatusCode)
	}
	defer resp.Body.Close()
	bytes_body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	doh_response = new(DOHResponse)
	err = json.Unmarshal(bytes_body, doh_response)
	if err != nil {
		return nil, err
	}
	if doh_response.Answer == nil {
		return nil, xerrors.Errorf("No TXT result found for domain %s .", domain)
	}

	return doh_response, nil
}

func parseTxt(txtField string) (result TXTPayload, err error) {
	kv := make(map[string]string)

	// txtField = "\"ps:true;v:1;sig:3QgQUPrPiBloBev8uf1wyjpa4roK4xjN2OXeBpqQFYMOHFo+blMR0Ppyc/JVj0jtdLDGBTrOFdOJPMfvUXZkwAE=;ca:1664267795;uuid:80c98711-f4f6-43c7-b05c-8d86372f6131;prev:null\""
	lo.ForEach(strings.Split(strings.Trim(txtField, "\""), ";"), func(combined string, i int) {
		pair := strings.Split(combined, ":")
		if len(pair) != 2 {
			err = xerrors.Errorf("TXT payload format error in %s", combined)
			return
		}
		kv[pair[0]] = pair[1]
	})
	if err != nil {
		return TXTPayload{}, err
	}
	if !lo.Every(lo.Keys(kv), []string{"ps", "v", "sig", "ca", "uuid", "prev"}) {
		return TXTPayload{}, xerrors.New("TXT payload not recognized: field missing")
	}
	if kv["ps"] != "true" {
		return TXTPayload{}, xerrors.New("TXT payload not recognized")
	}
	if kv["v"] != "1" {
		return TXTPayload{}, xerrors.New("TXT payload version not recognized")
	}

	uuid, err := uuid.Parse(kv["uuid"])
	if err != nil {
		return TXTPayload{}, err
	}
	createdAt, err := util.TimestampStringToTime(kv["ca"])
	if err != nil {
		return TXTPayload{}, err
	}

	var previous *string = nil
	if kv["prev"] != "null" {
		prev := kv["prev"]
		previous = &prev
	}

	return TXTPayload{
		Version:   1,
		Signature: kv["sig"],
		CreatedAt: createdAt,
		uuid:      uuid,
		Previous:  previous,
	}, nil
}
