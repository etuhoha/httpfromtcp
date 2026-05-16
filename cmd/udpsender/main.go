package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

func main() {
	addr, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		log.Fatal("could not resolve UDP addr\n")
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatal("could not establish UDP connection\n")
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		str, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			log.Printf("failed to read from stdin: %v\n", err)
		}

		_, err = conn.Write([]byte(str))
		if err != nil {
			log.Printf("failed to send via UDP: %v\n", err)
		}
	}
}
