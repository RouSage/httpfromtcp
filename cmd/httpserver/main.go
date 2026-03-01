package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
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
	if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin") {
		proxyHandler(w, req)
		return
	}

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

	res := []byte(res200)

	hs.Set("Content-Length", fmt.Sprintf("%d", len(res)))
	w.WriteStatusLine(response.StatusOK)
	w.WriteHeaders(hs)
	w.WriteBody(res)
}

func proxyHandler(w *response.Writer, req *request.Request) {
	path := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin")
	url := "https://httpbin.org" + path

	resp, err := http.Get(url)
	if err != nil {
		hs := response.GetDefaultHeaders(len(res500))
		hs.Set("Content-Type", "text/html")
		w.WriteStatusLine(response.StatusInternalServerError)
		w.WriteHeaders(hs)
		w.WriteBody([]byte(res500))
		return
	}
	defer resp.Body.Close()

	w.WriteStatusLine(response.StatusOK)

	hs := response.GetDefaultHeaders(0)
	hs.Set("Transfer-Encoding", "chunked")
	delete(hs, "content-length")

	w.WriteHeaders(hs)

	buf := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			_, err = w.WriteChunkedBody(buf[:n])
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Error reading from httpbin: %v", err)
			break
		}
	}
	w.WriteChunkedBodyDone()
}
