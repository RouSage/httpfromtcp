package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func main() {
	f, err := os.Open("messages.txt")
	if err != nil {
		log.Fatal()
	}
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
				fmt.Printf("read: %s\n", currLine)
				currLine = strings.Join(parts[1:], "")
			}
		}

		if errors.Is(err, io.EOF) {
			break
		}
	}

	if currLine != "" {
		fmt.Printf("read: %s\n", currLine)
	}
}
