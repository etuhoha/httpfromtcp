package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
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
	file, err := os.Open("messages.txt")
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	ch := getLinesChannel(file)
	for line := range ch {
		fmt.Printf("read: %v\n", line)
	}
}
