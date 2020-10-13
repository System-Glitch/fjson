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
	host     string
	port     int
	hostStr  string
	routines int
	duration int
)

func handleFlags() {
	flag.StringVar(&host, "host", "localhost", "Host")
	flag.IntVar(&port, "port", 8080, "Port")
	flag.IntVar(&routines, "routines", 1, "Number of goroutines")
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

	count := uint64(0)
	countChan := make(chan uint64, routines)
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*time.Duration(duration)))
	defer cancel()

	for i := 0; i < routines; i++ {
		go func() {
			count := uint64(0)
			for {
				select {
				case <-ctx.Done():
					countChan <- count
					return
				default:
					client.Send(payload)
					count++
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
	// TODO benchmark here
}
