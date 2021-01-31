package main

import (
	"log"
	"time"

	"github.com/System-Glitch/fjson"
)

func main() {
	client := fjson.NewClient("localhost:8080", time.Second)

	data := map[string]interface{}{
		"messages": []string{"hello", "world"},
	}
	resp, err := client.Send(data)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("%#v\n", resp)
}
