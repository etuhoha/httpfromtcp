package server

import (
	"fmt"
	"net"
	"sync/atomic"
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
	// request, err := request.RequestFromReader(conn)
	// if err != nil {
	// 	if s.closed.Load() {
	// 		return
	// 	}

	// 	log.Printf("error reading request: %v", err)
	// 	return
	// }

	// request
	resp := `HTTP/1.1 200 OK
Content-Type: text/plain
Content-Length: 13

Hello World!`

	conn.Write([]byte(resp))
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
