package rule

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

var (
	// domainModifierRegex matches domain modifier entries.
	//
	// The need for this regex comes from the fact that domain modifiers can contain regular expressions,
	// which can contain the separator character (|). This makes it impossible to just split the modifier by the separator.
	domainModifierRegex = regexp.MustCompile(`~?((/.*/)|[^|]+)+`)
)

type domainModifier struct {
	entries  []domainModifierEntry
	inverted bool
}

var _ matchingModifier = (*domainModifier)(nil)

func (m *domainModifier) Parse(modifier string) error {
	eqIndex := strings.IndexByte(modifier, '=')
	if eqIndex == -1 || eqIndex == len(modifier)-1 {
		return errors.New("invalid domain modifier")
	}
	value := modifier[eqIndex+1:]

	m.inverted = strings.HasPrefix(value, "~")
	matches := domainModifierRegex.FindAllString(value, -1)
	m.entries = make([]domainModifierEntry, len(matches))
	for i, entry := range matches {
		inverted := len(entry) > 0 && entry[0] == '~'
		if inverted != m.inverted {
			return errors.New("cannot mix inverted and non-inverted method modifiers")
		}
		if inverted {
			entry = entry[1:]
		}

		m.entries[i] = domainModifierEntry{}
		if err := m.entries[i].Parse(entry); err != nil {
			return fmt.Errorf("parse entry (%s): %w", entry, err)
		}
	}
	return nil
}

func (m *domainModifier) ShouldMatchReq(req *http.Request) bool {
	if referer := req.Header.Get("Referer"); referer == "" {
		return false
	}
	url, err := url.Parse(req.Header.Get("Referer"))
	if err != nil {
		return false
	}
	hostname := url.Hostname()

	matches := false
	for _, entry := range m.entries {
		if entry.MatchDomain(hostname) {
			matches = true
			break
		}
	}
	if m.inverted {
		return !matches
	}
	return matches
}

func (m *domainModifier) ShouldMatchRes(_ *http.Response) bool {
	return false
}

type domainModifierEntry struct {
	regular string
	tld     string
	regexp  *regexp.Regexp
}

func (m *domainModifierEntry) Parse(entry string) error {
	if len(entry) == 0 {
		return errors.New("entry is empty")
	}

	regexp, err := parseRegexp(entry)
	if err != nil {
		return fmt.Errorf("parse regexp: %w", err)
	}
	if regexp != nil {
		m.regexp = regexp
		return nil
	}

	if entry[len(entry)-1] == '*' {
		m.tld = entry[:len(entry)-1]
		return nil
	}

	m.regular = entry
	return nil
}

func (m *domainModifierEntry) MatchDomain(domain string) bool {
	switch {
	case m.regular != "":
		return strings.HasSuffix(domain, m.regular)
	case m.tld != "":
		return strings.HasPrefix(domain, m.tld)
	case m.regexp != nil:
		return m.regexp.MatchString(domain)
	default:
		return false
	}
}
