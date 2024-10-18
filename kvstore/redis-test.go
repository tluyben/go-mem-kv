package kvstore

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"github.com/gomodule/redigo/redis"
)

type RedisBenchmarkResult struct {
	Size           int
	ReadsPerSec    float64
	WritesPerSec   float64
	SearchesPerSec float64
	ScansPerSec    float64
	DeletesPerSec  float64
	UpdatesPerSec  float64
	ExistsPerSec   float64
	OpsPerSec      float64
	KBPerSec       float64
	MemoryUsageMB  float64
	Duration       time.Duration
	Incomplete     bool
}


func RunRedisBenchmark(address string, numRecords int, timeout time.Duration) RedisBenchmarkResult {
	pool := &redis.Pool{
		MaxIdle:     10,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", address)
		},
	}

	conn := pool.Get()
	defer conn.Close()

	keys := make([]string, numRecords)
	values := make([]string, numRecords)

	// Prepopulate with variable-sized values
	for i := 0; i < numRecords; i++ {
		keys[i] = fmt.Sprintf("key:%d:%s", i, randString(5, 10))
		values[i] = randString(50, 1000)
		_, err := conn.Do("SET", keys[i], values[i])
		if err != nil {
			fmt.Printf("Error setting key: %v\n", err)
		}
	}

	hotKeys := keys[:numRecords/5]

	var wg sync.WaitGroup
	// var mu sync.Mutex
	reads, writes, searches, scans, deletes, updates, exists := 0, 0, 0, 0, 0, 0, 0
	totalBytes := 0

	startTime := time.Now()
	done := make(chan struct{})
	timer := time.NewTimer(timeout)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			conn := pool.Get()
			defer conn.Close()

			localReads, localWrites, localSearches, localScans, localDeletes, localUpdates, localExists := 0, 0, 0, 0, 0, 0, 0
			localBytes := 0

			for {
				select {
				case <-done:
					return
				default:
					op := rand.Float32()
					var key string
					if rand.Float32() < 0.8 {
						key = hotKeys[rand.Intn(len(hotKeys))]
					} else {
						key = keys[rand.Intn(len(keys))]
					}

					simulateNetworkLatency()

					switch {
					case op < 0.45: // 45% reads
						_, err := conn.Do("GET", key)
						if err != nil {
							fmt.Printf("Error reading key: %v\n", err)
						}
						localReads++
					case op < 0.65: // 20% writes
						value := randString(50, 1000)
						_, err := conn.Do("SET", key, value)
						if err != nil {
							fmt.Printf("Error writing key: %v\n", err)
						}
						localWrites++
						localBytes += len(key) + len(value)
					case op < 0.75: // 10% searches
						_, err := conn.Do("KEYS", "*")
						if err != nil {
							fmt.Printf("Error searching keys: %v\n", err)
						}
						localSearches++
					case op < 0.85: // 10% scans
						cursor := 0
						for {
							values, err := redis.Values(conn.Do("SCAN", cursor))
							if err != nil {
								fmt.Printf("Error scanning: %v\n", err)
								break
							}
							cursor, _ = redis.Int(values[0], nil)
							if cursor == 0 {
								break
							}
						}
						localScans++
					case op < 0.90: // 5% deletes
						_, err := conn.Do("DEL", key)
						if err != nil {
							fmt.Printf("Error deleting key: %v\n", err)
						}
						localDeletes++
					case op < 0.95: // 5% updates
						value := randString(50, 1000)
						_, err := conn.Do("SET", key, value)
						if err != nil {
							fmt.Printf("Error updating key: %v\n", err)
						}
						localUpdates++
						localBytes += len(value)
					default: // 5% exists
						_, err := conn.Do("EXISTS", key)
						if err != nil {
							fmt.Printf("Error checking key existence: %v\n", err)
						}
						localExists++
					}
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(done)
	}()

	var incomplete bool
	select {
	case <-timer.C:
		incomplete = true
	case <-done:
	}

	duration := time.Since(startTime)
	totalOps := reads + writes + searches + scans + deletes + updates + exists

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return RedisBenchmarkResult{
		Size:           numRecords,
		ReadsPerSec:    float64(reads) / duration.Seconds(),
		WritesPerSec:   float64(writes) / duration.Seconds(),
		SearchesPerSec: float64(searches) / duration.Seconds(),
		ScansPerSec:    float64(scans) / duration.Seconds(),
		DeletesPerSec:  float64(deletes) / duration.Seconds(),
		UpdatesPerSec:  float64(updates) / duration.Seconds(),
		ExistsPerSec:   float64(exists) / duration.Seconds(),
		OpsPerSec:      float64(totalOps) / duration.Seconds(),
		KBPerSec:       float64(totalBytes) / 1024 / duration.Seconds(),
		MemoryUsageMB:  float64(m.Alloc) / (1024 * 1024),
		Duration:       duration,
		Incomplete:     incomplete,
	}
}