package main

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/etuhoha/httpfromtcp/internal/headers"
	"github.com/etuhoha/httpfromtcp/internal/request"
	"github.com/etuhoha/httpfromtcp/internal/response"
	"github.com/etuhoha/httpfromtcp/internal/server"
)

const port = 42069

const htmlTemplate = `<html>
  <head>
    <title>%d %s</title>
  </head>
  <body>
    <h1>%s</h1>
    <p>%s</p>
  </body>
</html>
`

const HTTPBIN_PREFIX = "/httpbin"

const BUFFER_SIZE = 1024

func main() {
	server, err := server.Serve(port, func(w *response.Writer, req *request.Request) {
		reqTarget := req.RequestLine.RequestTarget
		if after, ok := strings.CutPrefix(reqTarget, HTTPBIN_PREFIX); ok {
			err := handleHttpBin(w, after)
			if err != nil {
				w.WriteError(500, err)
			}
			return
		}

		if reqTarget == "/video" {
			err := handleVideo(w)
			if err != nil {
				w.WriteError(500, err)
			}
			return
		}

		statusCode := response.StatusCode(response.StatusOK)
		title := "Success!"
		text := "Your request was an absolute banger."

		switch req.RequestLine.RequestTarget {
		case "/yourproblem":
			statusCode = 400
			title = response.StatusCodeReason(statusCode)
			text = "Your request honestly kinda sucked."
		case "/myproblem":
			statusCode = 500
			title = response.StatusCodeReason(statusCode)
			text = "Okay, you know what? This one is on me."
		}

		msg := fmt.Sprintf(htmlTemplate, statusCode, response.StatusCodeReason(statusCode), title, text)
		w.WriteStatusLine(statusCode)

		headers := response.GetDefaultHeaders(len(msg))
		headers.Override("Content-Type", "text/html")
		w.WriteHeaders(headers)

		w.WriteBody([]byte(msg))
	})

	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	server.Close()
	log.Println("Server gracefully stopped")
}

func handleVideo(w *response.Writer) error {
	data, err := os.ReadFile("./assets/vim.mp4")
	if err != nil {
		return err
	}

	hdrs := response.GetDefaultHeaders(0)
	hdrs.Remove("Content-Length")
	hdrs.Set("Transfer-Encoding", "chunked")
	hdrs.Set("Trailer", "X-Content-Length")
	hdrs.Set("Trailer", "X-Content-SHA256")
	w.WriteHeaders(hdrs)

	return handleChunked(w, bytes.NewReader(data), "video/mp4")
}

func handleHttpBin(w *response.Writer, target string) error {
	resp, err := http.Get("https://httpbin.org/" + target)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 300 {
		return fmt.Errorf("could not get a proper result from httpbin.org, original status: %q", resp.Status)
	}

	respBody := resp.Body
	defer respBody.Close()
	return handleChunked(w, respBody, "application/json")
}

func handleChunked(w *response.Writer, reader io.Reader, contentType string) error {
	w.WriteStatusLine(200)

	hdrs := response.GetDefaultHeaders(0)
	hdrs.Remove("Content-Length")
	hdrs.Override("Content-Type", contentType)
	hdrs.Set("Transfer-Encoding", "chunked")
	hdrs.Set("Trailer", "X-Content-Length")
	hdrs.Set("Trailer", "X-Content-SHA256")
	w.WriteHeaders(hdrs)

	body := make([]byte, 0)

	buf := make([]byte, BUFFER_SIZE)
	for {
		n, err := reader.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		if n == 0 {
			break
		}

		body = append(body, buf[:n]...)

		_, err = w.WriteChunkedBody(buf[:n])
		if err != nil {
			return err
		}
	}

	sha := sha256.Sum256(body)

	trailHdrs := headers.NewHeaders()
	trailHdrs.Set("X-Content-Length", fmt.Sprintf("%d", len(body)))
	trailHdrs.Set("X-Content-SHA256", fmt.Sprintf("%x", sha))
	w.WriteChunkedBodyDone(&trailHdrs)
	return nil
}
