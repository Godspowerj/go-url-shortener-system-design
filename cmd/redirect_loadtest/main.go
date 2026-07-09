package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

const targetURL = "http://localhost:8080"

type runStats struct {
	concurrency     int
	totalRequests   int
	successfulReqs  int64
	failedReqs      int64
	totalDuration   time.Duration
	throughput      float64
	avgLatencyMs    float64
	p50Ms           int64
	p95Ms           int64
	p99Ms           int64
	statusCodes     map[int]int64
	errors          map[string]int64
}

func main() {
	totalReqsOpt := flag.Int("requests", 1000, "Total requests per run")
	flag.Parse()

	fmt.Println("🚀 Initializing Redirect Load Test Suite...")
	fmt.Printf("🎯 Target server: %s\n", targetURL)
	fmt.Printf("📊 Requests per run: %d\n\n", *totalReqsOpt)

	// Step 1: Create a valid short URL to test redirection
	shortCode, err := createTestURL()
	if err != nil {
		fmt.Printf("❌ Failed to create test URL: %v\n", err)
		return
	}
	fmt.Printf("✅ Created test short code: '%s'\n\n", shortCode)

	concurrencies := []int{10, 20, 50, 100, 200, 500}
	var results []runStats

	for _, c := range concurrencies {
		fmt.Printf("🏃 Running test with Concurrency = %d... ", c)
		stats := runLoadTest(c, *totalReqsOpt, shortCode)
		results = append(results, stats)
		fmt.Println("Done.")
		// Small sleep between runs to let SQLite/OS settle
		time.Sleep(1 * time.Second)
	}

	printMarkdownReport(results)
}

func createTestURL() (string, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	payload := map[string]string{
		"url": "https://www.youtube.com/watch?v=rickroll-loadtest",
	}
	jsonPayload, _ := json.Marshal(payload)

	resp, err := client.Post(targetURL+"/shorten", "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected status: %d, body: %s", resp.StatusCode, string(body))
	}

	var responseData struct {
		ShortCode string `json:"short_code"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return "", err
	}

	return responseData.ShortCode, nil
}

func runLoadTest(concurrency int, totalReqs int, shortCode string) runStats {
	var successCount int64
	var failureCount int64
	var totalDurationMs int64

	var mu sync.Mutex
	statusCounts := make(map[int]int64)
	errCounts := make(map[string]int64)
	latencies := make([]int64, 0, totalReqs)

	jobs := make(chan int, totalReqs)
	for i := 0; i < totalReqs; i++ {
		jobs <- i
	}
	close(jobs)

	// Configure client with custom transport to avoid HTTP socket exhaustion
	transport := &http.Transport{
		MaxIdleConns:        concurrency * 2,
		MaxIdleConnsPerHost: concurrency * 2,
		MaxConnsPerHost:     0,
		IdleConnTimeout:     90 * time.Second,
	}

	client := &http.Client{
		Timeout:   5 * time.Second,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Do NOT follow redirects, we want to measure the 302 response itself
			return http.ErrUseLastResponse
		},
	}

	start := time.Now()
	var wg sync.WaitGroup

	for w := 0; w < concurrency; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range jobs {
				reqStart := time.Now()
				resp, err := client.Get(targetURL + "/" + shortCode)
				reqDuration := time.Since(reqStart).Milliseconds()

				if err != nil {
					atomic.AddInt64(&failureCount, 1)
					mu.Lock()
					errCounts[err.Error()]++
					mu.Unlock()
					continue
				}

				resp.Body.Close()

				mu.Lock()
				statusCounts[resp.StatusCode]++
				latencies = append(latencies, reqDuration)
				mu.Unlock()

				// 302 Redirect is the expected success status code
				if resp.StatusCode == http.StatusFound {
					atomic.AddInt64(&successCount, 1)
					atomic.AddInt64(&totalDurationMs, reqDuration)
				} else {
					atomic.AddInt64(&failureCount, 1)
				}
			}
		}()
	}

	wg.Wait()
	testDuration := time.Since(start)

	totalRequests := successCount + failureCount
	rps := float64(totalRequests) / testDuration.Seconds()

	var avgLatency float64
	if successCount > 0 {
		avgLatency = float64(totalDurationMs) / float64(successCount)
	}

	// Calculate percentiles
	var p50, p95, p99 int64
	if len(latencies) > 0 {
		sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
		pick := func(p float64) int64 {
			idx := int(p * float64(len(latencies)-1))
			return latencies[idx]
		}
		p50 = pick(0.50)
		p95 = pick(0.95)
		p99 = pick(0.99)
	}

	return runStats{
		concurrency:     concurrency,
		totalRequests:   totalReqs,
		successfulReqs:  successCount,
		failedReqs:      failureCount,
		totalDuration:   testDuration,
		throughput:      rps,
		avgLatencyMs:    avgLatency,
		p50Ms:           p50,
		p95Ms:           p95,
		p99Ms:           p99,
		statusCodes:     statusCounts,
		errors:          errCounts,
	}
}

func printMarkdownReport(results []runStats) {
	fmt.Println("\n## Benchmark Report: Redirect Endpoint (GET /:code)")
	fmt.Println("\n| Concurrency | Total Requests | Success (302) | Failed | RPS (Throughput) | Avg Latency | p50 | p95 | p99 |")
	fmt.Println("|---|---|---|---|---|---|---|---|---|")
	for _, r := range results {
		fmt.Printf("| %d | %d | %d | %d | %.2f req/s | %.2f ms | %d ms | %d ms | %d ms |\n",
			r.concurrency, r.totalRequests, r.successfulReqs, r.failedReqs, r.throughput, r.avgLatencyMs, r.p50Ms, r.p95Ms, r.p99Ms)
	}

	fmt.Println("\n### Status Code Breakdown")
	for _, r := range results {
		fmt.Printf("*   **Concurrency %d**:\n", r.concurrency)
		for code, count := range r.statusCodes {
			fmt.Printf("    *   Status %d: %d\n", code, count)
		}
		for errMsg, count := range r.errors {
			fmt.Printf("    *   Error (%s): %d\n", errMsg, count)
		}
	}
}
