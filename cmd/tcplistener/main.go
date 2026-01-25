package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

func main() {
	ln, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Accepted connection")

		for line := range getLinesChannel(conn) {
			fmt.Println(line)
		}

		log.Println("Connection closed")
	}
}

func getLinesChannel(f io.ReadCloser) <-chan string {
	lines := make(chan string)

	go func() {
		defer close(lines)
		defer f.Close()

		txt := make([]byte, 8)
		currLine := ""

		for {
			n, err := f.Read(txt)
			if n > 0 {
				parts := strings.Split(string(txt), "\n")
				if len(parts) == 1 {
					currLine = currLine + parts[0]
				} else if len(parts) > 1 {
					currLine = currLine + parts[0]
					lines <- currLine
					currLine = strings.Join(parts[1:], "")
				}
			}

			if err != nil {
				if errors.Is(err, io.EOF) {
					if len(currLine) > 0 {
						lines <- currLine
					}
				}
				break
			}
		}
	}()

	return lines
}
