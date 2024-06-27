package rule

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

type headerModifier struct {
	// name is the name of the header.
	name string
	// exact is non-empty when the modifier should match an exact header value.
	exact string
	// regexp is non-nil when the modifier should match a header value using a regular expression.
	regexp *regexp.Regexp
}

var _ matchingModifier = (*headerModifier)(nil)

func (h *headerModifier) Parse(modifier string) error {
	if len(modifier) == 0 {
		return errors.New("empty modifier")
	}

	eqIndex := strings.IndexByte(modifier, '=')
	if eqIndex == -1 || eqIndex == len(modifier)-1 {
		return errors.New("modifier must contain a specifier")
	}
	specifier := modifier[eqIndex+1:]

	switch split := strings.Split(specifier, ":"); len(split) {
	case 1:
		h.name = split[0]
	case 2:
		h.name = split[0]
		value := split[1]
		regexp, err := parseRegexp(value)
		if err != nil {
			return fmt.Errorf("parse regexp: %w", err)
		}
		if regexp != nil {
			h.regexp = regexp
			break
		}
		h.exact = value
	default:
		return errors.New("invalid specifier")
	}

	return nil
}

func (h *headerModifier) ShouldMatchReq(_ *http.Request) bool {
	return false
}

func (h *headerModifier) ShouldMatchRes(res *http.Response) bool {
	value := res.Header.Get(h.name)
	if len(value) == 0 {
		return false
	}

	if h.exact != "" {
		if value != h.exact {
			return false
		}
	}
	if h.regexp != nil {
		if !h.regexp.MatchString(value) {
			return false
		}
	}

	return true
}
