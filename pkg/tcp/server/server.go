package server

import (
	"licklib/pkg/tag"
	"net"
)

const bufferSize = 1 << 20

type Server interface {
	Serve() error
	Close()
}

type server struct {
	t        *tag.Tag
	listener net.Listener
	address  string
}

func NewServer(name, address string) (Server, error) {
	return &server{t: tag.New(name, address), address: address}, nil
}

func (slf *server) Serve() error {
	listener, err := net.Listen("tcp", slf.address)
	if err != nil {
		return slf.t.Errorf("failed to start listening on address <%v>: %w", slf.address, err)
	}
	slf.listener = listener

	slf.t.Log("waiting for connection...")
	for {
		conn, err := listener.Accept()
		if err != nil {
			return slf.t.Errorf("failed to accept connection: %w", err)
		}
		slf.t.Log("connection established: local address <%v>, remote address <%v>.", conn.LocalAddr(), conn.RemoteAddr())
		go slf.serve(conn)
	}
}

func (slf *server) serve(conn net.Conn) {
	remotaAddress := conn.RemoteAddr().String()
	defer func() {
		if err := conn.Close(); err != nil {
			slf.t.Log("failed to close connection with <%v>: %v", remotaAddress, err)
			return
		}
		slf.t.Log("connection with <%v> closed", remotaAddress)
	}()

	buf := make([]byte, bufferSize)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			slf.t.Log("failed to read from <%v>: %v", remotaAddress, err)
			return
		}
		message := buf[:n]
		slf.t.Log("read %v bytes message from <%v>: %v", n, remotaAddress, string(message))

		answer := append([]byte("echo... "), message...)
		n, err = conn.Write(answer)
		if err != nil {
			slf.t.Log("failed to answer: %v", err)
			continue
		}
		slf.t.Log("sent %v bytes answer to <%v>: %v", n, remotaAddress, string(answer))
	}
}

func (slf *server) Close() {
	if err := slf.listener.Close(); err != nil {
		slf.t.Log("failed to close listener: %v", err)
		return
	}
	slf.t.Log("shutdown...")
}
