package kvstore

import (
	"flag"
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"testing"
	"time"
)

const (
	letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

var timeoutSeconds = flag.Int("timeout", 600, "Timeout in seconds for each benchmark run")

func randString(min, max int) string {
	n := rand.Intn(max-min+1) + min
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

type BenchmarkResult struct {
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
	Duration 	 time.Duration
	Incomplete     bool
}
func simulateNetworkLatency() {
	time.Sleep(time.Duration(rand.Intn(5)) * time.Millisecond)
}

func runBenchmark(b *testing.B, numRecords int, timeout time.Duration) BenchmarkResult {
	b.StopTimer()
	store := New()
	keys := make([]string, numRecords)
	values := make([]string, numRecords)

	// Prepopulate with variable-sized values
	for i := 0; i < numRecords; i++ {
		keys[i] = fmt.Sprintf("key:%d:%s", i, randString(5, 10))
		values[i] = randString(50, 1000)
		store.Set(keys[i], values[i])
	}

	hotKeys := keys[:numRecords/5]

	b.StartTimer()

	var wg sync.WaitGroup
	var mu sync.Mutex
	reads, writes, searches, scans, deletes, updates, exists := 0, 0, 0, 0, 0, 0, 0
	totalBytes := 0

	startTime := time.Now()
	done := make(chan struct{})
	timer := time.NewTimer(timeout)
	
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			localReads, localWrites, localSearches, localScans, localDeletes, localUpdates, localExists := 0, 0, 0, 0, 0, 0, 0
			localBytes := 0

			for j := 0; j < b.N/10; j++ {
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
						store.Get(key)
						localReads++
					case op < 0.65: // 20% writes
						value := randString(50, 1000)
						store.Set(key, value)
						localWrites++
						localBytes += len(key) + len(value)
					case op < 0.75: // 10% searches
						store.Keys()
						localSearches++
					case op < 0.85: // 10% scans
						cursor, _ := store.Scan(0, "key:*", 100, "")
						for cursor != 0 {
							cursor, _ = store.Scan(cursor, "key:*", 100, "")
						}
						localScans++
					case op < 0.90: // 5% deletes
						store.Del(key)
						localDeletes++
					case op < 0.95: // 5% updates
						value := randString(50, 1000)
						store.Set(key, value)
						localUpdates++
						localBytes += len(value)
					default: // 5% exists
						store.Exists(key)
						localExists++
					}
				}
			}

			mu.Lock()
			reads += localReads
			writes += localWrites
			searches += localSearches
			scans += localScans
			deletes += localDeletes
			updates += localUpdates
			exists += localExists
			totalBytes += localBytes
			mu.Unlock()
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

	return BenchmarkResult{
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

func BenchmarkKVStore(b *testing.B) {
	sizes := []int{1000, 10000, 500000}//, 100000, 1000000, 1000000000}
	results := make([]BenchmarkResult, len(sizes))
	timeout := time.Duration(*timeoutSeconds) * time.Second // Set timeout to 10 minutes

	for i, size := range sizes {
		b.Run(fmt.Sprintf("Size%d", size), func(b *testing.B) {
			results[i] = runBenchmark(b, size, timeout)
		})
	}

	// Print results in a tabular format
	fmt.Println("| Size | Reads/s | Writes/s | Searches/s | Scans/s | Deletes/s | Updates/s | Exists/s | Ops/s | KB/s | Memory (MB) | Duration    | Status |")
	fmt.Println("|------|---------|----------|------------|---------|-----------|-----------|----------|-------|------|-------------|-------------|--------|")
	for _, result := range results {
		status := "Complete"
		if result.Incomplete {
			status = "Incomplete"
		}
		fmt.Printf("| %d | %.2f | %.2f | %.2f | %.2f | %.2f | %.2f | %.2f | %.2f | %.2f | %.2f | %d | %s |\n",
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
			int64(result.Duration / time.Millisecond),
			status)
	}
}