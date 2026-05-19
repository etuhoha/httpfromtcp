package server

import (
	"bytes"
	"fmt"
	"io"
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

type Handler func(w io.Writer, req *request.Request) *HandlerError

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

			fmt.Printf("error acceptong connection: %v", err)
			continue
		}

		s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	var hErr *HandlerError
	request, err := request.RequestFromReader(conn)
	if err != nil {
		hErr = &HandlerError{StatusCode: response.StatusBadRequest, Message: err.Error()}
	}

	statusCode := response.StatusCode(response.StatusOK)
	var buffer bytes.Buffer

	if hErr == nil {
		hErr = s.handler(&buffer, request)
	}

	if hErr != nil {
		_, err := buffer.Write([]byte(hErr.Message))
		if err != nil {
			fmt.Printf("error writing error: %v", err)
			return
		}
		statusCode = hErr.StatusCode
	}

	writeResponse(conn, &buffer, statusCode)
}

func writeResponse(conn net.Conn, bodyBuf *bytes.Buffer, statusCode response.StatusCode) {
	err := response.WriteStatusLine(conn, response.StatusCode(statusCode))
	if err != nil {
		fmt.Printf("error writing status: %v", err)
		return
	}
	err = response.WriteHeaders(conn, response.GetDefaultHeaders(bodyBuf.Len()))
	if err != nil {
		fmt.Printf("error writing headers: %v", err)
		return
	}

	_, err = conn.Write(bodyBuf.Bytes())
	if err != nil {
		fmt.Printf("error writing body: %v", err)
		return
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
