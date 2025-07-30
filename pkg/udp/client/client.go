package client

import (
	"errors"
	"licklib/pkg/tag"
	"net"
)

const bufferSize = 1 << 20

type Client struct {
	t       *tag.Tag
	conn    net.Conn
	address string
}

func NewClient(name, address string) (*Client, error) {
	return &Client{t: tag.New(name, address), address: address}, nil
}

func (slf *Client) Init() error {
	conn, err := net.Dial("udp", slf.address)
	if err != nil {
		return slf.t.Errorf("failed to dial to <%v>: %v", slf.address, err)
	}
	slf.conn = conn

	slf.t.Log("start interacting: local address <%v>, remote address <%v>.", slf.conn.LocalAddr(), slf.conn.RemoteAddr())
	return nil
}

func (slf *Client) Send(message string) error {
	writtenBytesCount, err := slf.conn.Write([]byte(message))
	if err != nil {
		return slf.t.Errorf("failed to write message: %v", err)
	}

	slf.t.Log("wrote %v bytes message: %v", writtenBytesCount, message)
	return nil
}

func (slf *Client) Read() error {
	buf := make([]byte, bufferSize)
	for {
		n, err := slf.conn.Read(buf)
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				slf.t.Log("stop reading")
				return slf.t.Error(err)
			}
			slf.t.Log("failed to read: %v", err)
			continue
		}
		slf.t.Log("read %v bytes message: %v", n, string(buf[:n]))
	}
}

func (slf *Client) Close() {
	if err := slf.conn.Close(); err != nil {
		slf.t.Log("failed to close interaction: %v", err)
		return
	}
	slf.t.Log("shutdown...")
}
