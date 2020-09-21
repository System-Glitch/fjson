package fjson

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Server struct {
	done     chan struct{}
	listener net.Listener
	wg       sync.WaitGroup
	Timeout  time.Duration
	Handler  Handler
}

type Handler func(data interface{}) (interface{}, error)

func NewServer(timeout time.Duration, handler Handler) *Server {
	s := &Server{
		done:    make(chan struct{}, 1),
		Timeout: timeout,
		Handler: handler,
	}
	return s
}

func (s *Server) Listen(host string) error {
	l, err := net.Listen("tcp", host)
	if err != nil {
		return err
	}
	s.listener = l
	return nil
}

func (s *Server) Serve() {
	l := s.listener
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			select {
			case <-s.done:
				return
			default:
				continue
			}
		}
		s.wg.Add(1)
		go s.handleConnection(conn)
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

		resp, err := s.Handler(reqJSON)
		if err != nil {
			done <- fmt.Errorf("%w: %v", ErrHandler, err)
			return
		}

		b, err := json.Marshal(resp)
		if err != nil {
			done <- fmt.Errorf("%w: %v", ErrMarshal, err)
			return
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
	s.done <- struct{}{}
	s.listener.Close()
	s.wg.Wait()
}

func ListenAndServe(host string, timeout time.Duration, handler Handler) {

	s := NewServer(timeout, handler)

	if err := s.Listen(host); err != nil {
		log.Println(err)
		return
	}

	sigChannel := make(chan os.Signal, 1)
	signal.Notify(sigChannel, syscall.SIGINT, syscall.SIGTERM)
	go s.Serve()

	<-sigChannel
	s.Shutdown()
}
