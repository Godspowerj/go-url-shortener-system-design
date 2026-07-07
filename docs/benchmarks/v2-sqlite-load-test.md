# Benchmark: v2 SQLite Persistence Under Load

- **Date:** 2026-07-07
- **Storage Backend:** SQLite (`modernc.org/sqlite` pure-Go driver)
- **Environment:** Localhost, Single Process API server

---

## Benchmark 1: SQLite (Default Configuration) - *Before Tuning*

### Load Test Configuration
- **Total Requests:** 1,000 (HTTP `POST /shorten`)
- **Concurrency:** 20 concurrent workers

### Results
- **Successful Requests:** 15 / 1,000 (1.5%)
- **Failed Requests:** 985 / 1,000 (98.5%)
- **Throughput:** ~125.72 req/sec (RPS)
- **Average Latency:** 153.82 ms

### Analysis & Diagnostic
SQLite defaulted to synchronous file-locking. Concurrent writes collided, causing a `"database is locked"` SQL error. 98.5% of requests returned `500 Internal Server Error`.

---

## Benchmark 2: SQLite (WAL Mode + 5s Timeout) - *Tuned (100 Concurrency)*

### Load Test Configuration
- **Total Requests:** 1,000 (HTTP `POST /shorten`)
- **Concurrency:** 100 concurrent workers

### Results
- **Successful Requests:** 1,000 / 1,000 (100%)
- **Failed Requests:** 0 / 1,000 (0%)
- **Throughput:** ~258.31 req/sec (RPS)
- **Average Latency:** 343.83 ms

---

## Benchmark 3: SQLite (WAL Mode + 5s Timeout) - *Tuned (200 Concurrency)*

### Load Test Configuration
- **Total Requests:** 1,000 (HTTP `POST /shorten`)
- **Concurrency:** 200 concurrent workers

### Results
- **Successful Requests:** 1,000 / 1,000 (100%)
- **Failed Requests:** 0 / 1,000 (0%)
- **Throughput:** ~278.10 req/sec (RPS)
- **Average Latency:** 604.11 ms

---

## Summary Analysis
Tuning SQLite with `journal_mode(WAL)` and `busy_timeout(5000)` completely resolved our write failures:
*   **Errors dropped to 0%** even under 200 concurrent connections.
*   **Throughput doubled** from ~125 RPS to ~278 RPS.
*   **Latency Trade-off:** Average latency increased under 200 concurrency (604.11 ms) because requests are safely queuing up inside SQLite's write lock rather than failing immediately. This is our target bottleneck for future optimization (Redis/PostgreSQL).
