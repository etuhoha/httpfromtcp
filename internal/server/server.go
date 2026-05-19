package server

import (
	"fmt"
	"net"
	"sync/atomic"

	"github.com/etuhoha/httpfromtcp/internal/request"
	"github.com/etuhoha/httpfromtcp/internal/response"
)

type Server struct {
	Port int

	handler  Handler
	listener net.Listener
	closed   atomic.Bool
}

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

type Handler func(w *response.Writer, req *request.Request)

func Serve(port int, hanler Handler) (*Server, error) {
	server := &Server{handler: hanler, closed: atomic.Bool{}}
	server.Port = port

	addrStr := fmt.Sprintf("127.0.0.1:%d", port)
	listener, err := net.Listen("tcp", addrStr)
	if err != nil {
		return nil, err
	}

	server.listener = listener

	go server.listen()

	return server, nil
}

func (s *Server) listen() {
	for !s.closed.Load() {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed.Load() {
				return
			}

			fmt.Printf("error accepting connection: %v\n", err)
			continue
		}

		s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	resWriter := response.NewResponseWriter(conn)

	request, err := request.RequestFromReader(conn)
	if err != nil {
		resWriter.WriteError(400, err)
		return
	} else {
		s.handler(resWriter, request)
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
