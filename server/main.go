package main

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"log"
	"net"
	"time"
)

func scanPack(data []byte, atEOF bool) (advance int, token []byte, err error) {
	// Scan until 0, marking end of pack.
	for i := 0; i < len(data); i++ {
		if data[i] == 0 {
			return i, data[:i], nil
		}
	}

	if atEOF {
		return len(data), data, io.ErrUnexpectedEOF
	}

	// Request more data.
	return 0, nil, nil
}

func main() {
	// TODO cleanup (use a server struct)
	// TODO graceful shutdown
	l, err := net.Listen("tcp", "127.0.0.1:8080")

	if err != nil {
		log.Fatal(err)
	}

	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go func(c net.Conn) {
			defer c.Close()

			done := make(chan error)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			go func() {
				reader := bufio.NewScanner(conn)
				reader.Split(scanPack)
				if !reader.Scan() {
					log.Println("scan failed") // TODO better error handling
				}

				var reqJSON interface{}
				if err := json.Unmarshal(reader.Bytes(), &reqJSON); err != nil {
					done <- err
					return
				}
				log.Printf("%#v\n", reqJSON)

				b, err := json.Marshal("world")
				if err != nil {
					done <- err
				}
				if _, err := c.Write(b); err != nil {
					done <- err
					return
				}
				_, err = c.Write([]byte{0})
				done <- err
			}()

			select {
			case <-ctx.Done():
				log.Println("connection closed early: timeout")
				return
			case err := <-done:
				if err != nil {
					log.Println(err)
				}
			}
		}(conn)
	}
}
