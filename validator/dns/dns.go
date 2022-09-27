package dns

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/nextdotid/proof_server/types"
	"github.com/nextdotid/proof_server/validator"
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
	CD        bool          `json:"CD"`
	Question  []DOHQuestion `json:"Question"`
	// If Answer is empty, this field will appear.
	Authority *[]DOHAnswer  `json:"Authority"`
	Answer    *[]DOHAnswer  `json:"Answer"`
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

type DNS struct {
	*validator.Base
}

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
	return map[string]string{} // TODO
}

func (dns *DNS) GenerateSignPayload() (payload string) {
	return "" // TODO
}

func (dns *DNS) Validate() (err error) {
	return nil // TODO
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
