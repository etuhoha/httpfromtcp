package headers

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

type Headers map[string]string

func NewHeaders() Headers {
	return make(Headers)
}

const crlf = "\r\n"

func (h Headers) Get(key string) string {
	return h[strings.ToLower(key)]
}

func (h Headers) Set(key string, value string) {
	key = strings.ToLower(key)
	if v, ok := h[key]; ok {
		value = v + ", " + value
	}
	h[key] = value
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	crlfIdx := bytes.Index(data, []byte(crlf))
	if crlfIdx < 0 { // not enough data
		return 0, false, nil
	}

	if crlfIdx == 0 { // end of headers
		return 2, true, nil
	}

	hdrStr := string(data[:crlfIdx])
	comps := strings.SplitN(hdrStr, ":", 2)
	if len(comps) < 2 {
		return 0, false, fmt.Errorf("No separator in the header: %q", hdrStr)
	}

	name := comps[0]
	err = validateHeaderName(name)
	if err != nil {
		return 0, false, err
	}
	name = strings.ToLower(name)

	value := strings.TrimSpace(comps[1])

	h.Set(name, value)
	return crlfIdx + 2, false, nil
}

func validateHeaderName(name string) error {
	trimmedName := strings.TrimSpace(name)
	if len(name) != len(trimmedName) {
		return fmt.Errorf("Unexpected spaces around header name: %q", name)
	}

	ok, err := regexp.MatchString("^[a-zA-Z0-9!#\\$%&'\\*\\+-\\.\\^_`|~]+$", name)
	if err != nil {
		return err
	}

	if !ok {
		return fmt.Errorf("Unexpected characters in header name: %q", name)
	}

	return nil
}
