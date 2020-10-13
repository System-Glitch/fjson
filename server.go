package fjson

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type Server struct {
	done     chan struct{}
	listener net.Listener
	wg       sync.WaitGroup
	Handler  Handler
}

type Handler func(data interface{}) (interface{}, error)

func NewServer(handler Handler) *Server {
	s := &Server{
		done:    make(chan struct{}, 1),
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

	for {
		reader := bufio.NewScanner(c)
		reader.Split(scanPack)
		if !reader.Scan() {
			// log.Println(ErrScan)
			return
		}
		var reqJSON interface{}
		if err := json.Unmarshal(reader.Bytes(), &reqJSON); err != nil {
			log.Println(fmt.Errorf("%w: %v", ErrUnmarshal, err))
			return
		}

		resp, err := s.Handler(reqJSON)
		if err != nil {
			log.Println(fmt.Errorf("%w: %v", ErrHandler, err))
			return
		}

		b, err := json.Marshal(resp)
		if err != nil {
			log.Println(fmt.Errorf("%w: %v", ErrMarshal, err))
			return
		}
		if _, err := c.Write(append(b, 0)); err != nil {
			log.Println(fmt.Errorf("%w: %v", ErrWrite, err))
			return
		}
	}
}

func (s *Server) Shutdown() {
	s.done <- struct{}{}
	s.listener.Close()
	s.wg.Wait()
}

func ListenAndServe(host string, handler Handler) {

	s := NewServer(handler)

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
