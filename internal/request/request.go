package request

import (
	"fmt"
	"io"
	"regexp"
	"strings"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	var result Request

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	_, reqLine, err := parseRequestLine(data)
	if err != nil {
		return nil, err
	}
	result.RequestLine = reqLine

	return &result, nil
}

func parseRequestLine(data []byte) (int, RequestLine, error) {
	rl := ""
	rlEnd := -1
	for i, b := range data {
		if b == '\r' {
			if len(data) == (i+1) || data[i+1] != '\n' {
				return -1, RequestLine{}, fmt.Errorf("no LF found after CR at %d", i)
			}

			rl = string(data[:i])
			rlEnd = i + 1
			break
		}
	}

	if rlEnd == -1 {
		return -1, RequestLine{}, fmt.Errorf("no CR found")
	}

	comps := strings.Split(rl, " ")
	if len(comps) != 3 {
		return -1, RequestLine{}, fmt.Errorf("expected 3 components of request line, got %d (%v)", len(comps), comps)
	}

	// method
	method := comps[0]
	ok, _ := regexp.MatchString("^[A-Z]+$", method)
	if !ok {
		return -1, RequestLine{}, fmt.Errorf("expected method containing only capital letters, got %q", method)
	}

	// request target
	requestTarget := comps[1]
	err := validateRequestTarget(requestTarget)
	if err != nil {
		return -1, RequestLine{}, err
	}

	// httpVersion
	if comps[2] != "HTTP/1.1" {
		return -1, RequestLine{}, fmt.Errorf("expected HTTP version to be 'HTTP/1.1', got %q", comps[2])
	}

	return rlEnd, RequestLine{Method: method, RequestTarget: requestTarget, HttpVersion: "1.1"}, nil
}

func validateRequestTarget(requestTarget string) error {
	if requestTarget == "*" {
		return nil
	}

	return nil
}
