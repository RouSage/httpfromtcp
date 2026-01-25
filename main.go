package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
)

func main() {
	f, err := os.Open("messages.txt")
	if err != nil {
		log.Fatal()
	}
	defer f.Close()

	txt := make([]byte, 8)

	for {
		n, err := f.Read(txt)
		if n > 0 {
			fmt.Printf("read: %s\n", txt)
		}

		if errors.Is(err, io.EOF) {
			break
		}
	}
}
