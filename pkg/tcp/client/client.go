package client

import (
	"licklib/pkg/tag"
	"net"
	"time"
)

const (
	bufferSize      = 1 << 20
	reconnectPeriod = time.Second
)

type Client struct {
	t       *tag.Tag
	conn    net.Conn
	address string
}

func NewClient(name, address string) (*Client, error) {
	return &Client{t: tag.New(name, address), address: address}, nil
}

func (slf *Client) Connect(attempts int) error {
	for i := 1; i <= attempts; i++ {
		slf.t.Log("attempt %v/%v to connect...", i, attempts)
		conn, err := net.Dial("tcp", slf.address)
		if err != nil {
			slf.t.Log("failed to connect to <%v>: %v", slf.address, err)
			time.Sleep(reconnectPeriod)
			continue
		}
		slf.conn = conn
		break
	}

	if slf.conn == nil {
		return slf.t.Errorf("failed to connect to <%v>", slf.address)
	}

	slf.t.Log("connection established: local address <%v>, remote address <%v>.", slf.conn.LocalAddr(), slf.conn.RemoteAddr())
	return nil
}

func (slf *Client) Send(message string) error {
	n, err := slf.conn.Write([]byte(message))
	if err != nil {
		return slf.t.Errorf("failed to write to connection: %v", err)
	}

	slf.t.Log("wrote %v bytes to connection: <%v>", n, message)
	return nil
}

func (slf *Client) Read() error {
	buf := make([]byte, bufferSize)
	for {
		n, err := slf.conn.Read(buf)
		if err != nil {
			return slf.t.Errorf("failed to read from connection: %v", err)
		}
		slf.t.Log("read %v bytes message from connection: %v", n, string(buf[:n]))
	}
}

func (slf *Client) Close() {
	if err := slf.conn.Close(); err != nil {
		slf.t.Log("failed to close connection: %v", err)
		return
	}
	slf.t.Log("shutdown...")
}
