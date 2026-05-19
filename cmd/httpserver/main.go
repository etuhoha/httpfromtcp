package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

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

const httpBinPrefix = "/httpbin"

func main() {
	server, err := server.Serve(port, func(w *response.Writer, req *request.Request) {
		reqTarget := req.RequestLine.RequestTarget
		if after, ok := strings.CutPrefix(reqTarget, httpBinPrefix); ok {
			err := handleHttpBin(w, after)
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

func handleHttpBin(w *response.Writer, target string) error {
	resp, err := http.Get("https://httpbin.org/" + target)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 300 {
		return fmt.Errorf("could not get a proper result from httpbin.org, original status: %q", resp.Status)
	}

	w.WriteStatusLine(200)

	headers := response.GetDefaultHeaders(0)
	headers.Remove("Content-Length", "")
	headers.Set("Transfer-Encoding", "chunked")
	w.WriteHeaders(headers)

	reader := resp.Body

	buf := make([]byte, 32)
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

		w.WriteChunkedBody(buf[:n])
	}

	w.WriteChunkedBodyDone()
	return nil
}
