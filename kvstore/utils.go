package kvstore

import (
	"math/rand"
	"time"
)

const (
	letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

func randString(min, max int) string {
	n := rand.Intn(max-min+1) + min
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
func simulateNetworkLatency() {
	time.Sleep(time.Duration(rand.Intn(5)) * time.Millisecond)
}