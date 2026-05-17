package request

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
)

type Request struct {
	RequestLine RequestLine
	parseStatus parseStatus
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type parseStatus int

const (
	Init = iota
	Done
)

const BUFFER_SIZE = 80

func RequestFromReader(reader io.Reader) (*Request, error) {
	// fmt.Printf("START\n")
	result := Request{parseStatus: Init}

	buf := make([]byte, BUFFER_SIZE)
	readN := 0
	for result.parseStatus != Done {
		if readN == len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
			// fmt.Printf("resizing to size %v, new buf: %q\n", len(buf), string(buf[:readN]))
		}

		n, err := reader.Read(buf[readN:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				if n == 0 {
					// fmt.Printf("unexpected EOF\n")
					return nil, fmt.Errorf("unexpected EOF: %v", err)
				}
			} else {
				return nil, err
			}
		}
		readN += n
		// fmt.Printf("read %v, new buf: %q\n", n, string(buf[:readN]))

		parsedN, err := result.parse(buf[:readN])
		if err != nil {
			// fmt.Printf("ERR: %v\n", err)
			return nil, err
		}

		if parsedN != 0 {
			copy(buf, buf[parsedN:readN])
			readN -= parsedN
			// fmt.Printf("PARSED %v, new buf: %q\n", parsedN, string(buf[:readN]))
		}
	}

	return &result, nil
}

func (r *Request) parse(data []byte) (int, error) {
	switch r.parseStatus {
	case Init:
		parsedN, reqLine, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}

		if parsedN > 0 {
			r.RequestLine = reqLine
			r.parseStatus = Done
			return parsedN, nil
		}
	case Done:
		return 0, fmt.Errorf("unexpected parse in Done state")
	default:
		return 0, fmt.Errorf("unexpected parse status: %v", r.parseStatus)
	}

	return 0, nil
}

func parseRequestLine(data []byte) (int, RequestLine, error) {
	rl := ""
	readN := -1
	for i, b := range data {
		if b == '\n' {
			return -1, RequestLine{}, fmt.Errorf("unexpected LF at %d", i)
		}

		if b == '\r' {
			if len(data) == (i + 1) { // not enough data for now
				break
			}

			if data[i+1] != '\n' {
				return -1, RequestLine{}, fmt.Errorf("no LF found after CR at %d", i)
			}

			rl = string(data[:i])
			readN = i + 2
			break
		}
	}

	if readN == -1 {
		return 0, RequestLine{}, nil
	}

	comps := strings.Split(rl, " ")
	if len(comps) != 3 {
		return 0, RequestLine{}, fmt.Errorf("expected 3 components of request line, got %d (%v)", len(comps), comps)
	}

	// method
	method := comps[0]
	ok, _ := regexp.MatchString("^[A-Z]+$", method)
	if !ok {
		return 0, RequestLine{}, fmt.Errorf("expected method containing only capital letters, got %q", method)
	}

	// request target
	requestTarget := comps[1]
	err := validateRequestTarget(requestTarget)
	if err != nil {
		return 0, RequestLine{}, err
	}

	// httpVersion
	if comps[2] != "HTTP/1.1" {
		return 0, RequestLine{}, fmt.Errorf("expected HTTP version to be 'HTTP/1.1', got %q", comps[2])
	}

	return readN, RequestLine{Method: method, RequestTarget: requestTarget, HttpVersion: "1.1"}, nil
}

func validateRequestTarget(requestTarget string) error {
	if requestTarget == "*" {
		return nil
	}

	return nil
}
