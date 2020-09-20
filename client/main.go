package main

import (
	"log"
	"time"
)

func main() {
	client := NewClient("localhost:8080", 5*time.Second)

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