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

type writeState int

const (
	WriteStateStatusLine = iota
	WriteStateHeaders
	WriteStateBody
)

type ResponseWriter struct {
	writer     io.Writer
	writeState writeState
}

func StatusCodeReason(statusCode StatusCode) string {
	reasonPhrase := ""
	switch statusCode {
	case StatusOK:
		reasonPhrase = "OK"
	case StatusBadRequest:
		reasonPhrase = "Bad Request"
	case StatusInternalServerError:
		reasonPhrase = "Internal Server Error"
	}
	return reasonPhrase
}

func (w *ResponseWriter) WriteStatusLine(statusCode StatusCode) error {
	if w.writeState != WriteStateStatusLine {
		return fmt.Errorf("unexpected status line, expected: %v", writeStateString(w.writeState))
	}

	reasonPhrase := StatusCodeReason(statusCode)
	if reasonPhrase == "" {
		return fmt.Errorf("unknown status: %d", statusCode)
	}

	statusLine := fmt.Sprintf("HTTP/1.1 %d %s\r\n", statusCode, reasonPhrase)
	_, err := w.writer.Write([]byte(statusLine))
	if err != nil {
		return err
	}

	w.writeState = WriteStateHeaders
	return nil
}

func (w *ResponseWriter) WriteHeaders(headers headers.Headers) error {
	if w.writeState != WriteStateHeaders {
		return fmt.Errorf("unexpected headers, expected: %v", writeStateString(w.writeState))
	}

	for k, v := range headers {
		hStr := fmt.Sprintf("%s: %s\r\n", k, v)
		_, err := w.writer.Write([]byte(hStr))
		if err != nil {
			return err
		}
	}
	_, err := w.writer.Write([]byte("\r\n"))
	if err != nil {
		return err
	}

	w.writeState = WriteStateBody
	return nil
}

func (w *ResponseWriter) WriteBody(body []byte) error {
	if w.writeState != WriteStateBody {
		return fmt.Errorf("unexpected body, expected: %v", writeStateString(w.writeState))
	}

	_, err := w.writer.Write(body)
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

func NewResponseWriter(w io.Writer) *ResponseWriter {
	return &ResponseWriter{writer: w, writeState: WriteStateStatusLine}
}

func writeStateString(ws writeState) string {
	switch ws {
	case WriteStateStatusLine:
		return "status line"
	case WriteStateHeaders:
		return "headers"
	case WriteStateBody:
		return "body"
	default:
		return "unknown"
	}
}
