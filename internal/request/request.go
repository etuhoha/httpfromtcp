package request

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/etuhoha/httpfromtcp/internal/headers"
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte
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
	Header
	Body
	Done
)

const BUFFER_SIZE = 8

func RequestFromReader(reader io.Reader) (*Request, error) {
	// fmt.Printf("START\n")
	result := Request{parseStatus: Init}
	result.Headers = headers.NewHeaders()

	buf := make([]byte, BUFFER_SIZE)
	readN := 0
	lastParsed := false
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
				if !lastParsed {
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

		lastParsed = false
		if parsedN != 0 {
			lastParsed = true
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
			r.parseStatus = Header
			return parsedN, nil
		}
	case Header:
		parsedN, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}

		if done {
			r.parseStatus = Body
		}

		return parsedN, nil
	case Body:
		bodySize, err := parseBodySize(r.Headers)
		fmt.Printf("BSize %v, new buf: %q\n", bodySize, string(data))
		if err != nil {
			return 0, err
		}

		if bodySize == 0 || len(data) >= bodySize {
			r.Body = make([]byte, bodySize)
			copy(r.Body, data)
			fmt.Printf("BSize FIT! body: %q\n", string(r.Body))
			r.parseStatus = Done
			return bodySize, nil
		}

		return 0, nil
	case Done:
		return 0, fmt.Errorf("unexpected parse in Done state")
	default:
		return 0, fmt.Errorf("unexpected parse status: %v", r.parseStatus)
	}

	return 0, nil
}

func parseBodySize(headers headers.Headers) (int, error) {
	sizeStr, ok := headers["content-length"]
	if !ok {
		return 0, nil
	}

	size, err := strconv.Atoi(sizeStr)
	if err != nil {
		return 0, err
	}

	return size, nil
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
