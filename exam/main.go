package main

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"log"
	"os"
	"time"

	"github.com/System-Glitch/fjson"
)

type Response struct {
	Signatures [][]byte
	PubKey     crypto.PublicKey
}

var (
	curve = elliptic.P256()
	keys  *ecdsa.PrivateKey
)

func init() {
	var err error
	keys, err = ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Println("Could not generate key: ", err)
		os.Exit(1)
	}
}

func sign(data []byte) []byte {
	hash := sha256.Sum256(data)
	r, s, err := ecdsa.Sign(rand.Reader, keys, hash[:])
	if err != nil {
		log.Println("Signature error:", err)
		return nil
	}

	return append(r.Bytes(), s.Bytes()...)
}

func signAll(messages []string) *Response {
	resp := &Response{
		Signatures: make([][]byte, 0, len(messages)),
		PubKey:     keys.Public(),
	}
	for _, v := range messages {
		resp.Signatures = append(resp.Signatures, sign([]byte(v)))
	}
	return resp
}

func signAllParallel(messages []string) *Response {
	resp := &Response{
		Signatures: make([][]byte, 0, len(messages)),
		PubKey:     keys.Public(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
	defer cancel()
	res := make(chan []byte)
	routineCount := 64
	length := len(messages)
	if length < routineCount {
		routineCount = length
	}
	sliceLength := length / routineCount

	for i := 0; i < length; i += sliceLength {
		end := i + sliceLength
		if end > length {
			end = length
		}
		go func(messages []string) {
			for _, v := range messages {
				res <- sign([]byte(v))
			}
		}(messages[i:end])
	}

	for i := 0; i < routineCount; i++ {
		select {
		case <-ctx.Done():
			log.Println("Parallel hash timeout")
			return nil
		case sig := <-res:
			if sig == nil {
				continue
			}
			resp.Signatures = append(resp.Signatures, sig)
		}
	}

	return resp
}

func validateRequest(data interface{}) ([]string, bool) {
	obj, ok := data.(map[string]interface{})
	if !ok {
		return nil, false
	}

	m, ok := obj["messages"]
	if !ok {
		return nil, false
	}

	messages, ok := m.([]interface{})
	if !ok {
		return nil, false
	}

	res := make([]string, 0, len(messages))
	for _, v := range messages {
		str, ok := v.(string)
		if !ok {
			return nil, false
		}
		res = append(res, str)
	}

	return res, true
}

func main() {
	fjson.ListenAndServe("127.0.0.1:8080", func(data interface{}) (interface{}, error) {
		messages, ok := validateRequest(data)
		if !ok {
			return map[string]string{"error": "Malformed request."}, nil
		}
		return signAllParallel(messages), nil
	})
}
