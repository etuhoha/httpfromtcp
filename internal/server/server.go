package server

import (
	"fmt"
	"net"
	"sync/atomic"

	"github.com/etuhoha/httpfromtcp/internal/response"
)

type Server struct {
	Port int

	listener net.Listener
	closed   atomic.Bool
}

func Serve(port int) (*Server, error) {
	server := Server{closed: atomic.Bool{}}
	server.Port = port

	addrStr := fmt.Sprintf("127.0.0.1:%d", port)
	listener, err := net.Listen("tcp", addrStr)
	if err != nil {
		return nil, err
	}

	server.listener = listener

	go server.listen()

	return &server, nil
}

func (s *Server) listen() {
	for !s.closed.Load() {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed.Load() {
				return
			}

			continue
		}

		s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	err := response.WriteStatusLine(conn, 200)
	if err != nil {
		fmt.Printf("error writing status: %v", err)
	}
	err = response.WriteHeaders(conn, response.GetDefaultHeaders(0))
	if err != nil {
		fmt.Printf("error writing headers: %v", err)
	}
}

func (s *Server) Close() error {
	if s.closed.Load() {
		return fmt.Errorf("already closed")
	}

	err := s.listener.Close()
	if err != nil {
		return err
	}

	s.closed.Store(true)
	return nil
}
