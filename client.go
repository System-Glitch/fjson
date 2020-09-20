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

func NewClient(host string, timeout time.Duration) *Client {
	return &Client{
		Host:    host,
		Timeout: timeout,
	}
}

func (c *Client) Send(body interface{}) (interface{}, error) {
	var d net.Dialer
	data := make(chan interface{}, 1)
	errChan := make(chan error, 1)
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	conn, err := d.DialContext(ctx, "tcp", c.Host)

	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDial, err)
	}
	defer conn.Close()
	go func() {

		b, err := json.Marshal(body)
		if err != nil {
			errChan <- fmt.Errorf("%w: %v", ErrMarshal, err)
			return
		}

		if _, err := conn.Write(append(b, 0)); err != nil {
			errChan <- fmt.Errorf("%w: %v", ErrWrite, err)
			return
		}

		reader := bufio.NewScanner(conn)
		reader.Split(scanPack)
		if !reader.Scan() {
			errChan <- fmt.Errorf("%w: %v", ErrRead, err)
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
