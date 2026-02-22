package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/rousage/httpfromtcp/internal/request"
	"github.com/rousage/httpfromtcp/internal/response"
	"github.com/rousage/httpfromtcp/internal/server"
)

const (
	port   = 42069
	res400 = `<html>
        <head>
            <title>400 Bad Request</title>
        </head>
        <body>
            <h1>Bad Request</h1>
            <p>Your request honestly kinda sucked.</p>
        </body>
        </html>
`
	res500 = `<html>
        <head>
            <title>500 Internal Server Error</title>
        </head>
        <body>
            <h1>Internal Server Error</h1>
            <p>Okay, you know what? This one is on me.</p>
        </body>
        </html>
`
	res200 = `<html>
        <head>
            <title>200 OK</title>
        </head>
        <body>
            <h1>Success!</h1>
            <p>Your request was an absolute banger.</p>
        </body>
        </html>
`
)

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handler(w *response.Writer, req *request.Request) {
	hs := response.GetDefaultHeaders(0)
	hs.Set("Content-Type", "text/html")

	if req.RequestLine.RequestTarget == "/yourproblem" {
		hs.Set("Content-Length", fmt.Sprintf("%d", len(res400)))
		w.WriteStatusLine(response.StatusBadRequest)
		w.WriteHeaders(hs)
		w.WriteBody([]byte(res400))
		return
	}
	if req.RequestLine.RequestTarget == "/myproblem" {
		hs.Set("Content-Length", fmt.Sprintf("%d", len(res500)))
		w.WriteStatusLine(response.StatusInternalServerError)
		w.WriteHeaders(hs)
		w.WriteBody([]byte(res500))
		return
	}

	err := w.WriteStatusLine(response.StatusOK)
	if err != nil {
		hs.Set("Content-Length", fmt.Sprintf("%d", len(res500)))
		w.WriteStatusLine(response.StatusInternalServerError)
		w.WriteHeaders(hs)
		w.WriteBody([]byte(res500))
		return
	}

	res := []byte(res200)

	hs.Set("Content-Length", fmt.Sprintf("%d", len(res)))
	w.WriteStatusLine(response.StatusOK)
	w.WriteHeaders(hs)
	w.WriteBody(res)
}
