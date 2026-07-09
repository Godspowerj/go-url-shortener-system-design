# ADR-004: Asynchronous Click Tracking via Buffered Channels

- **Status:** Proposed
- **Date:** 2026-07-09

## Context
During load testing of our redirection endpoint (`GET /:code`), we discovered that our throughput was capped and latency spiked significantly under high concurrency (e.g., average latency over **413 ms** at 100 concurrency, and severe timeouts at 200+ concurrency). 

The root cause is the synchronous execution of `UPDATE click_count` inside the request path. Because SQLite only allows one write transaction at a time, concurrent HTTP requests are forced to serialize and queue up waiting for the SQLite write lock. This blocks Go routines and saturates the OS listen backlog, causing the server to refuse incoming connections.

## Decision
We will decouple database write transactions from the HTTP request-response cycle by introducing an **Asynchronous Background Worker** backed by a Go **Buffered Channel**.

*   When a short URL is resolved, the service will publish the URL ID to a buffered channel (`clicksQueue chan int`).
*   The handler will immediately receive the URL and redirect the user without waiting for the database write to complete.
*   A background goroutine (worker) will start on application launch, consume URL IDs from the channel, and update the click counts sequentially in SQLite.

## Alternatives Considered

1.  **Synchronous Writes (Current status):** Insufficient due to SQLite serialization bottlenecks under moderate load.
2.  **Redis Caching/Counter:** Rejected for now to avoid introducing external infrastructure dependencies. We want a pure Go lightweight solution.
3.  **Spinning up a goroutine per request (`go s.store.Update(...)`):** Rejected because it bypasses write serialization. Hundreds of concurrent goroutines concurrently attempting to write to SQLite would cause severe `"database is locked"` errors.

## Pros and Cons

*   **Pros:**
    *   **Sub-millisecond redirection:** Redirects return immediately (no database write waiting).
    *   **Serialized Database Writes:** The background worker writes sequentially, eliminating SQLite lock contention entirely.
    *   **Lightweight:** Built using standard Go channels and goroutines; no external dependencies required.
*   **Cons:**
    *   **Eventual Consistency:** The database click counts will update slightly after the redirect completes (milliseconds delay).
    *   **In-Memory Buffer Risk:** If the server crashes abruptly, any click events sitting in the buffered channel that haven't been written to disk will be lost.

## Consequences
*   We accept eventual consistency for url analytics.
*   We must choose a buffer limit for the channel. If the buffer is full, we will handle it gracefully (e.g., using a non-blocking select to log an error and drop the update rather than blocking the web request).

## Validation
*   Run the redirect benchmark suite up to 500 concurrency.
*   Confirm throughput increases significantly and failed request timeouts drop to 0.
