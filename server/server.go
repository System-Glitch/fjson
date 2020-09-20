package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var (
	ErrScan      error = errors.New("FJSON Request scanning failed")
	ErrMarshal   error = errors.New("FJSON marshal error")
	ErrWrite     error = errors.New("FJSON write error")
	ErrUnmarshal error = errors.New("FJSON unmarshal error")
)

type Server struct {
	stopping chan bool
	newConn  chan net.Conn
	wg       sync.WaitGroup
	Timeout  time.Duration
}

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

func NewServer(timeout time.Duration) *Server {
	return &Server{
		newConn:  make(chan net.Conn, 1),
		stopping: make(chan bool, 1),
		Timeout:  timeout,
	}
}

func (s *Server) Listen(host string) {
	l, err := net.Listen("tcp", host)

	if err != nil {
		log.Println(err)
	}

	defer l.Close()

	for {
		go func() {
			conn, err := l.Accept()
			if err != nil {
				return
			}
			s.newConn <- conn
		}()
		select {
		case <-s.stopping:
			return
		case conn := <-s.newConn:
			s.wg.Add(1)
			go s.handleConnection(conn)
		}
	}
}

func (s *Server) handleConnection(c net.Conn) {
	defer s.wg.Done()
	defer c.Close()

	done := make(chan error)
	ctx, cancel := context.WithTimeout(context.Background(), s.Timeout)
	defer cancel()

	go func() {
		reader := bufio.NewScanner(c)
		reader.Split(scanPack)
		if !reader.Scan() {
			done <- ErrScan
			return
		}

		var reqJSON interface{}
		if err := json.Unmarshal(reader.Bytes(), &reqJSON); err != nil {
			done <- fmt.Errorf("%w: %v", ErrUnmarshal, err)
			return
		}
		log.Printf("%#v\n", reqJSON)
		// TODO add custom handler
		b, err := json.Marshal("world")
		if err != nil {
			done <- fmt.Errorf("%w: %v", ErrMarshal, err)
		}
		if _, err := c.Write(append(b, 0)); err != nil {
			done <- fmt.Errorf("%w: %v", ErrWrite, err)
			return
		}
		done <- nil
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
}

func (s *Server) Shutdown() {
	s.stopping <- true
	s.wg.Wait()
}

func Listen(host string, timeout time.Duration) {

	s := NewServer(timeout)

	sigChannel := make(chan os.Signal, 1)
	signal.Notify(sigChannel, syscall.SIGINT, syscall.SIGTERM)

	go s.Listen(host)

	// No need for timeout checking, as each request can already timeout
	<-sigChannel // Block until SIGINT or SIGTERM received
	s.Shutdown()
}
