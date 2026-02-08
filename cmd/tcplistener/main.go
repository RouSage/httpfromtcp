package main

import (
	"fmt"
	"log"
	"net"

	"github.com/rousage/httpfromtcp/internal/request"
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

		req, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Request line:\n- Method: %s\n- Target: %s\n- Version: %s\n", req.RequestLine.Method, req.RequestLine.RequestTarget, req.RequestLine.HttpVersion)
		fmt.Printf("Headers:\n")
		for key, val := range req.Headers {
			fmt.Printf("- %s: %s\n", key, val)
		}
		fmt.Printf("Body:\n%s", req.Body)

		log.Println("Connection closed")
	}
}
