package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

const (
	targetURL   = "http://localhost:8080"
	concurrency = 500
	totalReqs   = 1000
)

func main() {
	fmt.Printf("🎯 Starting Load Test against %s...\n", targetURL)
	fmt.Printf("   Concurrency: %d concurrent workers\n", concurrency)
	fmt.Printf("   Total Requests: %d\n\n", totalReqs)

	var successCount int64
	var failureCount int64
	var totalDurationMs int64

	jobs := make(chan int, totalReqs)
	for i := 0; i < totalReqs; i++ {
		jobs <- i
	}
	close(jobs)

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	start := time.Now()
	var wg sync.WaitGroup

	for w := 0; w < concurrency; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range jobs {
				reqStart := time.Now()

				payload := map[string]string{
					"url": fmt.Sprintf("https://www.youtube.com/watch?v=loadtest-%d", time.Now().UnixNano()),
				}
				jsonPayload, _ := json.Marshal(payload)

				resp, err := client.Post(targetURL+"/shorten", "application/json", bytes.NewBuffer(jsonPayload))
				reqDuration := time.Since(reqStart).Milliseconds()
				atomic.AddInt64(&totalDurationMs, reqDuration)

				if err != nil || resp.StatusCode != http.StatusCreated {
					atomic.AddInt64(&failureCount, 1)
					if resp != nil {
						resp.Body.Close()
					}
					continue
				}

				resp.Body.Close()
				atomic.AddInt64(&successCount, 1)
			}
		}()
	}

	wg.Wait()
	testDuration := time.Since(start)

	totalRequests := successCount + failureCount
	rps := float64(totalRequests) / testDuration.Seconds()
	avgLatency := float64(totalDurationMs) / float64(totalRequests)

	fmt.Println("📊 --- Load Test Results ---")
	fmt.Printf("⏱️  Total Duration:     %v\n", testDuration)
	fmt.Printf("✅ Successful Req:     %d\n", successCount)
	fmt.Printf("❌ Failed Req:         %d\n", failureCount)
	fmt.Printf("⚡ Throughput:         %.2f req/sec (RPS)\n", rps)
	fmt.Printf("⏳ Average Latency:    %.2f ms\n", avgLatency)
}
