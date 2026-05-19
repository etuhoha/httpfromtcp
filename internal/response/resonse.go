package response

import (
	"fmt"
	"io"
	"strconv"

	"github.com/etuhoha/httpfromtcp/internal/headers"
)

type StatusCode int

const (
	StatusOK                  = 200
	StatusBadRequest          = 400
	StatusInternalServerError = 500
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	var reasonPhrase string
	switch statusCode {
	case StatusOK:
		reasonPhrase = "OK"
	case StatusBadRequest:
		reasonPhrase = "Bad Request"
	case StatusInternalServerError:
		reasonPhrase = "Internal Server Error"
	default:
		return fmt.Errorf("unknown status: %d", statusCode)
	}

	statusLine := fmt.Sprintf("HTTP/1.1 %d %s\r\n", statusCode, reasonPhrase)
	_, err := w.Write([]byte(statusLine))
	if err != nil {
		return err
	}

	return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	result := headers.NewHeaders()
	result.Set("Content-Length", strconv.FormatInt(int64(contentLen), 10))
	result.Set("Connection", "close")
	result.Set("Content-Type", "text/plain")
	return result
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for k, v := range headers {
		hStr := fmt.Sprintf("%s: %s\r\n", k, v)
		_, err := w.Write([]byte(hStr))
		if err != nil {
			return err
		}
	}
	_, err := w.Write([]byte("\r\n"))
	if err != nil {
		return err
	}

	return nil
}
