package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

const (
	targetURL   = "http://localhost:8080"
	concurrency = 100
	totalReqs   = 1000
)

type result struct {
	statusCode int
	err        error
	durationMs int64
	success    bool
}

func main() {
	fmt.Printf("🎯 Starting Load Test against %s...\n", targetURL)
	fmt.Printf("   Concurrency: %d concurrent workers\n", concurrency)
	fmt.Printf("   Total Requests: %d\n\n", totalReqs)

	var successCount int64
	var failureCount int64
	var successDurationMs int64
	var failureDurationMs int64

	// Guarded by mutex since map writes aren't atomic-safe.
	var statusMu sync.Mutex
	statusCounts := make(map[int]int64)
	errCounts := make(map[string]int64)

	// Keep a small sample of failure bodies for debugging, capped so we
	// don't spam the console or eat memory on high failure rates.
	var sampleMu sync.Mutex
	const maxSamples = 10
	failureSamples := make([]string, 0, maxSamples)

	// Track latencies for a p50/p95/p99 view rather than just an average.
	var latMu sync.Mutex
	successLatencies := make([]int64, 0, totalReqs)

	jobs := make(chan int, totalReqs)
	for i := 0; i < totalReqs; i++ {
		jobs <- i
	}
	close(jobs)

	// Default transport caps idle conns per host at 2, which causes
	// connection churn under high concurrency and can itself look like
	// server-side failure. Raise the limits so the client isn't the bottleneck.
	transport := &http.Transport{
		MaxIdleConns:        concurrency * 2,
		MaxIdleConnsPerHost: concurrency * 2,
		MaxConnsPerHost:     0, // unlimited
		IdleConnTimeout:     90 * time.Second,
	}

	client := &http.Client{
		Timeout:   5 * time.Second,
		Transport: transport,
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

				if err != nil {
					atomic.AddInt64(&failureCount, 1)
					atomic.AddInt64(&failureDurationMs, reqDuration)

					statusMu.Lock()
					errCounts[err.Error()]++
					statusMu.Unlock()

					addSample(&sampleMu, &failureSamples, maxSamples, fmt.Sprintf("err=%v", err))
					continue
				}

				body, _ := io.ReadAll(resp.Body)
				resp.Body.Close()

				statusMu.Lock()
				statusCounts[resp.StatusCode]++
				statusMu.Unlock()

				if resp.StatusCode != http.StatusCreated {
					atomic.AddInt64(&failureCount, 1)
					atomic.AddInt64(&failureDurationMs, reqDuration)
					addSample(&sampleMu, &failureSamples, maxSamples,
						fmt.Sprintf("status=%d body=%s", resp.StatusCode, truncate(string(body), 200)))
					continue
				}

				atomic.AddInt64(&successCount, 1)
				atomic.AddInt64(&successDurationMs, reqDuration)

				latMu.Lock()
				successLatencies = append(successLatencies, reqDuration)
				latMu.Unlock()
			}
		}()
	}

	wg.Wait()
	testDuration := time.Since(start)

	totalRequests := successCount + failureCount
	rps := float64(totalRequests) / testDuration.Seconds()

	var avgSuccessLatency, avgFailureLatency float64
	if successCount > 0 {
		avgSuccessLatency = float64(successDurationMs) / float64(successCount)
	}
	if failureCount > 0 {
		avgFailureLatency = float64(failureDurationMs) / float64(failureCount)
	}

	fmt.Println("📊 --- Load Test Results ---")
	fmt.Printf("⏱️  Total Duration:       %v\n", testDuration)
	fmt.Printf("✅ Successful Req:       %d\n", successCount)
	fmt.Printf("❌ Failed Req:           %d\n", failureCount)
	fmt.Printf("⚡ Throughput:           %.2f req/sec (RPS)\n", rps)
	fmt.Printf("⏳ Avg Success Latency:  %.2f ms\n", avgSuccessLatency)
	fmt.Printf("⏳ Avg Failure Latency:  %.2f ms\n", avgFailureLatency)

	if p := percentiles(successLatencies); p != nil {
		fmt.Printf("📈 Success Latency p50/p95/p99: %d / %d / %d ms\n", p[0], p[1], p[2])
	}

	if len(statusCounts) > 0 {
		fmt.Println("\n📋 --- Status Code Breakdown ---")
		codes := make([]int, 0, len(statusCounts))
		for code := range statusCounts {
			codes = append(codes, code)
		}
		sort.Ints(codes)
		for _, code := range codes {
			fmt.Printf("   %d: %d\n", code, statusCounts[code])
		}
	}

	if len(errCounts) > 0 {
		fmt.Println("\n📋 --- Client Error Breakdown ---")
		for msg, count := range errCounts {
			fmt.Printf("   %d x  %s\n", count, msg)
		}
	}

	if len(failureSamples) > 0 {
		fmt.Println("\n🔍 --- Sample Failures (first", len(failureSamples), ") ---")
		for _, s := range failureSamples {
			fmt.Println("  ", s)
		}
	}
}

func addSample(mu *sync.Mutex, samples *[]string, max int, s string) {
	mu.Lock()
	defer mu.Unlock()
	if len(*samples) < max {
		*samples = append(*samples, s)
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// percentiles returns [p50, p95, p99] in ms, or nil if there's no data.
func percentiles(latencies []int64) []int64 {
	if len(latencies) == 0 {
		return nil
	}
	sorted := make([]int64, len(latencies))
	copy(sorted, latencies)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })

	pick := func(p float64) int64 {
		idx := int(p * float64(len(sorted)-1))
		return sorted[idx]
	}

	return []int64{pick(0.50), pick(0.95), pick(0.99)}
}