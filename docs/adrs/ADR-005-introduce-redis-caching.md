# ADR-005: Introduce Redis Caching for Redirects

- **Status:** Proposed
- **Date:** 2026-07-10

## Context
As read queries (GET `/{code}`) scale up, querying SQLite directly on every redirect becomes a potential bottleneck due to database read locks and disk I/O, especially under concurrent write load. We need a fast in-memory caching layer to bypass SQLite for read requests of existing short links.

## Decision
We will introduce Redis as a memory cache for resolved URLs. The lookup flow for `ResolveURL` in [service.go](file:///c:/Users/Godspower%20jonah/Desktop/url-shortener/service/service.go) will be:
1. **Cache Look-up**: Check Redis for the short code.
2. **Cache Hit**: Deserialize the stored JSON string back into a `store.URL` struct and return it immediately.
3. **Cache Miss**: Query SQLite. If found, serialize the `store.URL` struct as JSON and store it in Redis with an expiration TTL (e.g., 24 hours). Return the URL.
4. **Click-Tracking**: Click increments will continue to run asynchronously using the existing background worker queue, which writes to SQLite.

To implement this, we will add `Get` and `Set` methods to `store.RedisStore` in [redis.go](file:///c:/Users/Godspower%20jonah/Desktop/url-shortener/store/redis.go).

## Alternatives Considered
- **In-Memory Go Map Cache**: Simple, but doesn't scale horizontally across multiple instances of the API server and lacks built-in TTL eviction.
- **SQLite Only (No Cache)**: Lowers architectural complexity but degrades API server performance under high read concurrency.

## Pros and Cons
- **Pros:**
  - Extremely fast read responses (sub-millisecond cache hits).
  - Reduced read load on the primary SQLite database.
  - Easy key-value access with TTL.
- **Cons:**
  - Increased infrastructure complexity (requires a running Redis instance).
  - Potential cache consistency edge cases if we support deletion/updates in the future (though we currently do not).

## Consequences
- The URL service now requires a valid Redis connection. If Redis is down, we should gracefully fall back to querying SQLite directly (fail-soft).

## Validation
- We will run the redirect load test ([cmd/redirect_loadtest/main.go](file:///c:/Users/Godspower%20jonah/Desktop/url-shortener/cmd/redirect_loadtest/main.go)) to measure read throughput and latency before and after implementing Redis caching.
