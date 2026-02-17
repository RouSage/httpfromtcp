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

func (he *HandlerError) Write(w io.Writer) {
	io.WriteString(w, fmt.Sprintf("HTTP/1.1 %d %s\r\n", he.StatusCode, he.Message))
}

type Server struct {
	listener net.Listener
	handler  Handler
	closed   atomic.Bool
}

func Serve(port int, handler Handler) (*Server, error) {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	s := &Server{
		listener: l,
		handler:  handler,
	}

	go s.listen()

	return s, nil
}

func (s *Server) Close() error {
	s.closed.Store(true)
	return s.listener.Close()
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed.Load() {
				return
			}
			log.Printf("Error accepting connection: %v", err)
			continue
		}

		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	req, err := request.RequestFromReader(conn)
	if err != nil {
		hErr := &HandlerError{StatusCode: response.StatusBadRequest, Message: err.Error()}
		hErr.Write(conn)
		return
	}

	buf := bytes.NewBuffer([]byte{})
	hErr := s.handler(buf, req)
	if hErr != nil {
		hErr.Write(conn)
		return
	}

	err = response.WriteStatusLine(conn, response.StatusOK)
	if err != nil {
		log.Fatal(err)
	}

	hs := response.GetDefaultHeaders(buf.Len())
	err = response.WriteHeaders(conn, hs)
	if err != nil {
		hErr := &HandlerError{StatusCode: response.StatusInternalServerError, Message: err.Error()}
		hErr.Write(conn)
		return
	}

	err = response.WriteBody(conn, buf.Bytes())
	if err != nil {
		hErr := &HandlerError{StatusCode: response.StatusInternalServerError, Message: err.Error()}
		hErr.Write(conn)
		return
	}
}
