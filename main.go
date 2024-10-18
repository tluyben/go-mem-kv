package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/tluyben/go-mem-kv/kvstore"
)

func main() {
	// Define the port flag
	port := flag.Int("port", 6379, "Port number to run the Redis-compatible server on")
	flag.Parse()

	// Set the REDIS_PORT environment variable
	os.Setenv("REDIS_PORT", strconv.Itoa(*port))

	// Create a new KVStore instance
	store := kvstore.New()

	// Create a new RedisServer instance
	server := kvstore.NewRedisServer(store)

	// Start the server
	fmt.Printf("Starting Redis-compatible server on port %d\n", *port)
	fmt.Printf("Use 'telnet localhost %d' to connect\n", *port)
	err := server.Start()
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}