package server

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"github.com/rousage/httpfromtcp/internal/response"
)

type Server struct {
	listener net.Listener
	closed   atomic.Bool
}

func Serve(port int) (*Server, error) {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	s := &Server{
		listener: l,
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
			log.Fatal(err)
		}

		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	err := response.WriteStatusLine(conn, response.StatusOK)
	if err != nil {
		log.Fatal(err)
	}

	hs := response.GetDefaultHeaders(0)
	err = response.WriteHeaders(conn, hs)
	if err != nil {
		log.Fatal(err)
	}
}
