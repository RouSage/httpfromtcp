package server

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"sync/atomic"

	"github.com/rousage/httpfromtcp/internal/request"
	"github.com/rousage/httpfromtcp/internal/response"
)

type Handler func(w io.Writer, req *request.Request) *HandlerError

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

type Server struct {
	listener net.Listener
	closed   atomic.Bool
}

func Serve(port int, handler Handler) (*Server, error) {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	s := &Server{
		listener: l,
	}

	go s.listen(handler)

	return s, nil
}

func (s *Server) Close() error {
	s.closed.Store(true)
	return s.listener.Close()
}

func (s *Server) listen(handler Handler) {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed.Load() {
				return
			}
			log.Fatal(err)
		}

		go s.handle(conn, handler)
	}
}

func (s *Server) handle(conn net.Conn, handler Handler) {
	defer conn.Close()

	req, err := request.RequestFromReader(conn)
	if err != nil {
		log.Fatal(err)
	}

	var b bytes.Buffer
	handlerError := handler(&b, req)
	if handlerError != nil {
		err := writeError(conn, handlerError)
		if err != nil {
			log.Fatal(err)
		}
	}

	err = response.WriteStatusLine(conn, response.StatusOK)
	if err != nil {
		log.Fatal(err)
	}

	hs := response.GetDefaultHeaders(b.Len())
	err = response.WriteHeaders(conn, hs)
	if err != nil {
		log.Fatal(err)
	}

	err = response.WriteBody(conn, b.Bytes())
	if err != nil {
		log.Fatal(err)
	}
}

func writeError(w io.Writer, handlerErr *HandlerError) error {
	_, err := io.WriteString(w, fmt.Sprintf("HTTP/1.1 %d %s\r\n", handlerErr.StatusCode, handlerErr.Message))
	return err
}
