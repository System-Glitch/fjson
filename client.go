package fjson

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"
)

type Client struct {
	Host    string
	Timeout time.Duration
}

type Connection struct {
	conn   net.Conn
	client *Client
}

func NewClient(host string, timeout time.Duration) *Client {
	return &Client{
		Host:    host,
		Timeout: timeout,
	}
}

func (c *Client) Connect() (*Connection, error) {
	var d net.Dialer
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()
	conn, err := d.DialContext(ctx, "tcp", c.Host)

	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDial, err)
	}

	return &Connection{conn, c}, nil
}

func (c *Connection) Send(body interface{}) (interface{}, error) {
	data := make(chan interface{}, 1)
	errChan := make(chan error, 1)
	ctx, cancel := context.WithTimeout(context.Background(), c.client.Timeout)
	defer cancel()

	go func() {

		var payload []byte
		if b, ok := body.([]byte); ok { // Don't marshal if content is already a byte slice
			payload = b
		} else {
			b, err := json.Marshal(body)
			if err != nil {
				errChan <- fmt.Errorf("%w: %v", ErrMarshal, err)
				return
			}
			payload = b
		}

		if _, err := c.conn.Write(append(payload, 0)); err != nil {
			errChan <- fmt.Errorf("%w: %v", ErrWrite, err)
			return
		}

		reader := bufio.NewScanner(c.conn)
		reader.Split(scanPack)
		if !reader.Scan() {
			errChan <- ErrRead
			return
		}

		var respJSON interface{}
		if err := json.Unmarshal(reader.Bytes(), &respJSON); err != nil {
			errChan <- fmt.Errorf("%w: %v", ErrUnmarshal, err)
			return
		}
		data <- respJSON
	}()

	select {
	case <-ctx.Done():
		return nil, ErrTimeout
	case err := <-errChan:
		return nil, err
	case resp := <-data:
		return resp, nil
	}
}

func (c *Connection) Close() error {
	return c.conn.Close()
}

func (c *Client) Send(body interface{}) (interface{}, error) {
	conn, err := c.Connect()
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	return conn.Send(body)
}
