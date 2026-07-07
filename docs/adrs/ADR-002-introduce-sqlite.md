# ADR-002: Introduce SQLite for Persistent Storage (v2)

- **Status:** Accepted
- **Date:** 2026-07-04

## Context
Under ADR-001, we stored URLs in memory. As expected, whenever the application restarts or crashes, all shortened URLs are deleted. We need permanent storage so links remain active forever.

## Decision
We chose **SQLite** as our persistent database using a pure-Go driver (`modernc.org/sqlite`). This keeps the app persistent while saving database tables to a local file (`urls.db`).

## Alternatives Considered
- **PostgreSQL / MySQL:** Rejected for now. They require setting up a separate server instance, configuring users, managing network overhead, and adding connection credentials, which is overkill for a single-node setup.
- **JSON file on disk:** Rejected because flat files suffer from file corruption under concurrent write access, and query speeds are slow.

## Pros and Cons
- **Pros:**
  - **Survives restarts:** Data is saved to a file (`urls.db`).
  - **No external dependencies:** Pure Go driver requires no CGo compiler or running databases on the host.
  - Full SQL capability (tables, indexes, constraints).
- **Cons:**
  - **Write Bottleneck:** SQLite locks the entire file during writes, limiting concurrent write scaling.
  - Cannot scale horizontally across multiple servers (the file lives on one server).

## Consequences
All links now survive server restarts. We accept the write performance bottleneck since current traffic levels are low. If we scale to multiple API servers in the future, we will have to migrate to a client-server database like PostgreSQL.

## Validation
Will be validated by writing a custom load test tool to hit the persistent SQLite storage under concurrent load.
