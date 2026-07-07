# ADR-003: Introduce Store Interface for Decoupling

- **Status:** Accepted
- **Date:** 2026-07-04

## Context
When migrating from memory store to SQLite, the HTTP handlers (`handler/url.go`) were tightly coupled to `*store.MemoryStore`. Changing backends required rewriting the handler logic, which violates the Dependency Inversion Principle.

## Decision
We introduced a unified `store.Store` interface in `store/store.go`:
```go
type Store interface {
    Save(shortCode string, url URL) error
    Get(shortCode string) (URL, bool)
    GetAll() []URL
}
```
The HTTP handler now interacts only with the interface, not the concrete implementation.

## Alternatives Considered
- **Direct Refactoring:** Rewriting the handler to use `*store.SQLiteStore` directly. Rejected because it would require rewrite effort again when we migrate to PostgreSQL or Redis in future versions.

## Pros and Cons
- **Pros:**
  - **Decoupling:** Handler doesn't care whether the database is SQLite, Memory, or Postgres.
  - **Swappability:** We can change backends in `main.go` with one line of code.
  - **Testability:** Makes it easy to mock the database for testing.
- **Cons:**
  - Small layer of abstraction overhead.

## Consequences
We can now swap storage engines effortlessly. This prepares the codebase for clean architectural extensions in v3.

## Validation
Verified that the compiler checks interface compatibility automatically at build time.
