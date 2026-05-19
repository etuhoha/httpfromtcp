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
	WriteStateTrailer
	WriteStateDone
)

type Writer struct {
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

func (w *Writer) WriteError(statusCode StatusCode, err error) {
	w.WriteStatusLine(400)
	errMsg := err.Error()
	headers := GetDefaultHeaders(len(errMsg))
	w.WriteHeaders(headers)
	w.WriteBody([]byte(errMsg))
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
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

func (w *Writer) WriteHeadersRaw(headers headers.Headers) error {
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

	return nil
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.writeState != WriteStateHeaders {
		return fmt.Errorf("unexpected headers, expected: %v", writeStateString(w.writeState))
	}

	err := w.WriteHeadersRaw(headers)
	if err != nil {
		return err
	}

	w.writeState = WriteStateBody
	return nil
}

func (w *Writer) WriteBody(body []byte) error {
	if w.writeState != WriteStateBody {
		return fmt.Errorf("unexpected body, expected: %v", writeStateString(w.writeState))
	}

	_, err := w.writer.Write(body)
	if err != nil {
		return err
	}

	w.writeState = WriteStateDone
	return nil
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	if w.writeState != WriteStateBody {
		return 0, fmt.Errorf("unexpected body, expected: %v", writeStateString(w.writeState))
	}

	chunkLen := len(p)
	chunckLenStr := strconv.FormatInt(int64(chunkLen), 16)

	n1, err := w.writer.Write([]byte(chunckLenStr + "\r\n"))
	if err != nil {
		return 0, err
	}

	n2, err := w.writer.Write(p)
	if err != nil {
		return 0, err
	}

	n3, err := w.writer.Write([]byte("\r\n"))
	if err != nil {
		return 0, err
	}

	w.writeState = WriteStateTrailer
	return n1 + n2 + n3, nil
}

func (w *Writer) WriteChunkedBodyDone(trailHdrs *headers.Headers) error {
	if w.writeState != WriteStateTrailer {
		return fmt.Errorf("unexpected trailers, expected: %v", writeStateString(w.writeState))
	}

	_, err := w.writer.Write([]byte("0\r\n"))
	if err != nil {
		return err
	}

	if trailHdrs != nil {
		err = w.WriteHeadersRaw(*trailHdrs)
		if err != nil {
			return err
		}
	} else {
		_, err = w.writer.Write([]byte("\r\n"))
		if err != nil {
			return err
		}
	}

	w.writeState = WriteStateDone
	return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	result := headers.NewHeaders()
	result.Set("Content-Length", strconv.FormatInt(int64(contentLen), 10))
	result.Set("Connection", "close")
	result.Set("Content-Type", "text/plain")
	return result
}

func NewResponseWriter(w io.Writer) *Writer {
	return &Writer{writer: w, writeState: WriteStateStatusLine}
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
