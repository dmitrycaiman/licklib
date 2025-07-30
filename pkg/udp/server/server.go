package server

import (
	"licklib/pkg/tag"
	"net"
)

const bufferSize = 1 << 20

type Server struct {
	t        *tag.Tag
	listener net.PacketConn
	address  string
}

func NewServer(name, address string) (*Server, error) {
	return &Server{t: tag.New(name, address), address: address}, nil
}

func (slf *Server) Serve() error {
	listener, err := net.ListenPacket("udp", slf.address)
	if err != nil {
		return slf.t.Errorf("failed to start listening on address <%v>: %w", slf.address, err)
	}
	slf.listener = listener

	slf.t.Log("waiting for data on <%v>...", listener.LocalAddr())
	buf := make([]byte, bufferSize)
	for {
		n, addr, err := listener.ReadFrom(buf)
		if err != nil {
			slf.t.Log("failed to read from <%v>: %v", addr.String(), err)
			continue
		}

		message := buf[:n]
		slf.t.Log("read %v bytes message from <%v>: %v", n, addr.String(), string(message))
		go slf.serve(message, addr)
	}
}

func (slf *Server) serve(message []byte, addr net.Addr) {
	answer := append([]byte("echo... "), message...)
	n, err := slf.listener.WriteTo(answer, addr)
	if err != nil {
		slf.t.Log("failed to answer: %v", err)
		return
	}
	slf.t.Log("sent %v bytes answer to <%v>: %v", n, addr.String(), string(answer))
}

func (slf *Server) Close() {
	if err := slf.listener.Close(); err != nil {
		slf.t.Log("failed to close listener: %v", err)
		return
	}
	slf.t.Log("shutdown...")
}
