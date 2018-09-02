package conn

import (
	"net"
)

type Client struct {
	addr string
	Conn net.Conn

	WriteHandler ConnWriteHandler
}

func (c *Client) Send(p []byte) (n int, err error) {
	if c.Conn == nil {
		if err = c.Dial(); err != nil {
			return
		}
	}

	return c.WriteHandler(c.Conn, p)
}

func (c *Client) Dial() error {
	conn, err := net.Dial("tcp", c.addr)
	if err != nil {
		return err
	}

	c.Conn = conn
	return nil
}

func NewClient(addr string) (*Client, error) {
	c := &Client{
		addr: addr,
	}

	return c, nil
}
