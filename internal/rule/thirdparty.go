package rule

import (
	"net/http"
	"strings"
)

// https://adguard.com/kb/general/ad-filtering/create-own-filters/#third-party-modifier
type thirdPartyModifier struct {
	inverted bool
}

func (m *thirdPartyModifier) Parse(modifier string) error {
	if modifier[0] == '~' {
		m.inverted = true
	}
	return nil
}

func (m *thirdPartyModifier) ShouldMatchReq(req *http.Request) bool {
	if req.Header.Get("Sec-Fetch-Site") == "cross-site" {
		return !m.inverted
	}

	referer := req.Header.Get("Referer")
	if referer == "" {
		return false
	}
	targetHost := req.Host
	refererURL, err := req.URL.Parse(referer)
	if err != nil {
		return false
	}
	refererHost := refererURL.Hostname()
	if strings.HasSuffix(refererHost, targetHost) {
		return m.inverted
	}
	return !m.inverted
}

func (m *thirdPartyModifier) ShouldMatchRes(_ *http.Response) bool {
	return false
}
