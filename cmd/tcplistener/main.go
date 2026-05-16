package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

const READ_SIZE = 8

func getLinesChannel(f io.ReadCloser) <-chan string {
	ch := make(chan string)

	go func() {
		defer close(ch)
		defer f.Close()

		data := make([]byte, READ_SIZE)

		cur_line := ""
		for {
			n, err := f.Read(data)
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				log.Printf("error reading file: %v\n", err)
				return
			}

			chunk := string(data[:n])
			chunks := strings.Split(chunk, "\n")

			for _, c := range chunks[:len(chunks)-1] {
				ch <- cur_line + c
				cur_line = ""
			}

			cur_line += chunks[len(chunks)-1]
		}

		if len(cur_line) > 0 {
			ch <- cur_line
		}
	}()

	return ch
}

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

		ch := getLinesChannel(conn)
		for line := range ch {
			fmt.Printf("%v\n", line)
		}

		fmt.Println("Connection closed.")
	}
}
