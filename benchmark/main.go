package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/System-Glitch/fjson"
)

var (
	host        string
	port        int
	hostStr     string
	routines    int
	connections int
	duration    int
)

func handleFlags() {
	flag.StringVar(&host, "host", "localhost", "Host")
	flag.IntVar(&port, "port", 8080, "Port")
	flag.IntVar(&routines, "routines", 1, "Number of goroutines")
	flag.IntVar(&connections, "connections", 400, "Number of open connections")
	flag.IntVar(&duration, "duration", 30, "Number of seconds for which the benchmark will run")
	flag.Parse()

	hostStr = fmt.Sprintf("%s:%d", host, port)
}

func main() {
	handleFlags()
	client := fjson.NewClient(hostStr, time.Second*2)

	data := map[string]interface{}{
		"greetings": "hello",
		"now":       time.Now().Unix(),
	}
	payload, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}

	connectionCount := connections / routines
	count := uint64(0)
	countChan := make(chan uint64, routines)
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*time.Duration(duration)))
	defer cancel()

	for i := 0; i < routines; i++ {
		go func() {
			conns := make([]*fjson.Connection, 0, connectionCount)
			for j := 0; j < connectionCount; j++ {
				newConn, err := client.Connect()
				if err != nil {
					log.Println(err)
					continue
				}
				defer newConn.Close()
				conns = append(conns, newConn)
			}
			count := uint64(0)
			for {
				select {
				case <-ctx.Done():
					countChan <- count
					return
				default:
					for _, conn := range conns {
						conn.Send(payload)
					}
					count += uint64(len(conns))
				}
			}
		}()
	}

	for i := 0; i < routines; i++ {
		if c, ok := <-countChan; ok {
			count += c
		}
	}
	close(countChan)
	log.Printf("Routines: %d | Duration: %ds | Count: %d | R/s: %d\n", routines, duration, count, count/uint64(duration))
}
