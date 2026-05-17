package main

import (
	"fmt"
	"log"
	"net"

	"github.com/etuhoha/httpfromtcp/internal/request"
)

func main() {
	listener, err := net.Listen("tcp", "127.0.0.1:42069")
	if err != nil {
		log.Fatalf("error listening: %v", err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("error connecting: %v", err)
		}

		fmt.Println("Connection accepted!")

		request, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatalf("error parsing request: %v", err)
		}

		fmt.Printf("Request line:\n")
		fmt.Printf("- Method: %v\n", request.RequestLine.Method)
		fmt.Printf("- Target: %v\n", request.RequestLine.RequestTarget)
		fmt.Printf("- Version: %v\n", request.RequestLine.HttpVersion)

		fmt.Println("Connection closed.")
	}
}
