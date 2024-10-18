package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/tluyben/go-mem-kv/kvstore"
)

func main() {
	// Define the port flag
	port := flag.Int("port", 6379, "Port number to run the Redis-compatible server on")
	redisTest := flag.String("redis-test", "", "Run Redis benchmark (format: localhost:port)")
	flag.Parse()

	if *redisTest != "" {
		runRedisBenchmark(*redisTest)
		return
	}
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

func runRedisBenchmark(address string) {
	sizes := []int{1000, 10000, 500000}
	results := make([]kvstore.RedisBenchmarkResult, len(sizes))
	timeout := 10 * time.Minute

	for i, size := range sizes {
		fmt.Printf("Running benchmark for size %d...\n", size)
		results[i] = kvstore.RunRedisBenchmark(address, size, timeout)
	}

	// Print results in a tabular format
	fmt.Println("| Size | Reads/s | Writes/s | Searches/s | Scans/s | Deletes/s | Updates/s | Exists/s | Ops/s | KB/s | Memory (MB) | Duration | Status |")
	fmt.Println("|------|---------|----------|------------|---------|-----------|-----------|----------|-------|------|-------------|----------|--------|")
	for _, result := range results {
		status := "Complete"
		if result.Incomplete {
			status = "Incomplete"
		}
		fmt.Printf("| %d | %.2f | %.2f | %.2f | %.2f | %.2f | %.2f | %.2f | %.2f | %.2f | %.2f | %d ms | %s |\n",
			result.Size,
			result.ReadsPerSec,
			result.WritesPerSec,
			result.SearchesPerSec,
			result.ScansPerSec,
			result.DeletesPerSec,
			result.UpdatesPerSec,
			result.ExistsPerSec,
			result.OpsPerSec,
			result.KBPerSec,
			result.MemoryUsageMB,
			int64(result.Duration/time.Millisecond),
			status)
	}
}