# Benchmark: v2 SQLite Persistence Under Load

- **Date:** 2026-07-06
- **Storage Backend:** SQLite (`modernc.org/sqlite` pure-Go driver)
- **Environment:** Localhost, Single Process API server

## Load Test Configuration
- **Total Requests:** 1,000 (HTTP `POST /shorten`)
- **Concurrency:** 20 concurrent workers

## Results
- **Successful Requests:** 15 / 1,000 (1.5%)
- **Failed Requests:** 985 / 1,000 (98.5%)
- **Throughput:** ~125.72 req/sec (RPS)
- **Average Latency:** 153.82 ms

## Analysis & Diagnostic
The load test revealed a massive failure rate (98.5% of requests returned `500 Internal Server Error`).

Looking at the server logs, we confirmed that under concurrent load, multiple threads attempted to write (`INSERT`) into the SQLite database file at the exact same millisecond. Since SQLite by default locks the entire file during database writes, these concurrent requests conflicted, returning a `"database is locked"` SQL error. Go returned `500` status codes for those failed saves.

### Key Bottleneck Identified
- **Write Concurrency Limit:** SQLite's single-writer architecture cannot handle high-concurrency writes without tuning (e.g. enabling WAL mode, setting busy timeouts, or using single-writer worker queues).

## Next Action
To resolve this bottleneck in the future, we need to:
1. Try enabling **WAL (Write-Ahead Logging)** mode in SQLite, which allows concurrent readers while one writer is writing.
2. Configure a **busy timeout** connection parameter so SQLite waits a small duration for the lock to release instead of failing immediately.
3. Eventually introduce **Redis Caching** (for reads) or switch to a production-grade client-server database like **PostgreSQL** to scale writes.
