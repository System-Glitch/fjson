package main

import (
	"log"
	"time"
)

func main() {
	Listen("127.0.0.1:8080", time.Second, func(data interface{}) (interface{}, error) {
		log.Printf("%#v\n", data)
		return "world", nil
	})
}
