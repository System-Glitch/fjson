package main

import (
	"github.com/System-Glitch/fjson"
)

func main() {
	fjson.ListenAndServe("127.0.0.1:8080", func(data interface{}) (interface{}, error) {
		// log.Printf("%#v\n", data)
		return "world", nil
	})
}
