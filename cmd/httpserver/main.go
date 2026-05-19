package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
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

func main() {
	server, err := server.Serve(port, func(w *response.ResponseWriter, req *request.Request) {
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
