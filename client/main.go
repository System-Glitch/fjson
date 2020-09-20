package main

import (
	"log"
	"time"

	"github.com/System-Glitch/fjson"
)

func main() {
	client := fjson.NewClient("localhost:8080", time.Second)

	data := map[string]interface{}{
		"greetings": "hello",
		"now":       time.Now().Unix(),
	}
	resp, err := client.Send(data)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("%#v\n", resp)
}
