# ADR-001: Use In-Memory Storage for MVP (v1)

- **Status:** Superseded (by ADR-002)
- **Date:** 2026-06-30

## Context
We need to build a proof-of-concept (MVP) for our URL shortener. We want to test HTTP routes, request handlers, and B62 generation as quickly as possible without adding infrastructure complexity like databases.

## Decision
We chose to store URLs in RAM using a standard Go `map[string]URL` protected by a `sync.RWMutex` to handle concurrent reads/writes safely.

## Alternatives Considered
- **SQLite / PostgreSQL:** Rejected because setting up schemas, migrations, and database connections would slow down the initial development of the HTTP routing engine.

## Pros and Cons
- **Pros:**
  - Extremely fast read/write speeds (nanoseconds).
  - Zero setup or installation required.
  - Very simple implementation.
- **Cons:**
  - **Data loss:** All data disappears as soon as the server restarts.
  - **RAM usage:** Scale is limited by server memory.

## Consequences
While this allowed us to build the MVP in a single day, it is completely unusable in production since restarting the server deletes all user links.

## Validation
Validated manually via `curl` requests confirming short URLs successfully save and redirect during the server's lifecycle.
