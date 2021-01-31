package main

import (
	"fmt"
	"testing"
)

func generateMessages() []string {
	messageCount := 10000
	messages := make([]string, 0, messageCount)
	for i := 0; i < messageCount; i++ {
		messages = append(messages, fmt.Sprintf("%d", i))
	}
	return messages
}

func BenchmarkSerial(b *testing.B) {
	messages := generateMessages()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		signAll(messages)
	}
}

func BenchmarkParallel(b *testing.B) {
	messages := generateMessages()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		signAllParallel(messages)
	}
}
