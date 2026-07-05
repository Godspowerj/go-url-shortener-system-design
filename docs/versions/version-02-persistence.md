# Version 2 — Persistence

## Problem

The current application stores all URLs in memory.

When the server stops or restarts, every shortened URL is lost.

This makes the application unreliable because users expect their links to remain available over time.

---

## Goal

Replace the in-memory storage with persistent storage so that URLs survive server restarts.

---

## Current Architecture (Version 1)

```mermaid
graph TD
    Client["Client (curl / browser)"]
    API["Go API (Chi Router)"]
    Store["Memory Store\n(Go map in RAM)"]

    Client -->|HTTP Request| API
    API -->|Save / Get / GetAll| Store
    Store -->|URL data| API
    API -->|HTTP Response| Client

    style Store fill:#f66,color:#fff
    style Store stroke:#c00
```

> ⚠️ RAM is wiped when the server stops. All data is lost.

---

## Target Architecture (Version 2)

```mermaid
graph TD
    Client["Client (curl / browser)"]
    API["Go API (Chi Router)"]
    DB["SQLite Database\n(urls.db file on disk)"]

    Client -->|HTTP Request| API
    API -->|SQL queries| DB
    DB -->|URL rows| API
    API -->|HTTP Response| Client

    style DB fill:#4c9,color:#fff
    style DB stroke:#090
```

> ✅ SQLite writes to a file on disk. Data survives restarts.

---

## Before vs After

```mermaid
graph LR
    subgraph Version1["Version 1 — Memory"]
        M["map[string]URL\n(lives in RAM)"]
    end

    subgraph Version2["Version 2 — SQLite"]
        D["urls.db\n(lives on disk)"]
    end

    Version1 -->|"replace store layer only"| Version2
```

The handler and router stay exactly the same.
Only the store layer is swapped.

---

## Why are we making this change?

Memory is temporary.

A database persists data even after the application stops running.

---

## Questions to Answer

* What is a database?
* Why SQLite?
* Why not PostgreSQL yet?
* What information do we need to store?
* How should our application communicate with SQLite?
* Can we replace our storage without changing the rest of the application?

---

## What Changes vs What Stays the Same

```mermaid
graph TD
    main["main.go"] 
    handler["handler/url.go"]
    shortener["shortener/generate.go"]
    memory["store/memory.go"]
    sqlite["store/sqlite.go"]

    main --> handler
    handler --> shortener
    handler --> sqlite

    style memory fill:#f66,color:#fff,stroke:#c00
    style sqlite fill:#4c9,color:#fff,stroke:#090
```

| File | Status |
|------|--------|
| `main.go` | ✅ No change |
| `handler/url.go` | ✅ No change |
| `shortener/generate.go` | ✅ No change |
| `store/memory.go` | 🔴 Replaced |
| `store/sqlite.go` | 🟢 New file |

---

## New Data Flow — POST /shorten

```mermaid
sequenceDiagram
    participant Client
    participant Handler as Shorten() (handler/url.go)
    participant Store as SQLiteStore (store/sqlite.go)
    participant DB as urls.db (SQLite file)

    Client->>Handler: POST /shorten {"url": "https://youtube.com"}
    Handler->>Handler: decode body, generate short code
    Handler->>Store: Save("xK92pQ", URL entry)
    Store->>DB: INSERT INTO urls VALUES (...)
    DB-->>Store: row saved ✓
    Store-->>Handler: done
    Handler-->>Client: 201 {"short_code": "xK92pQ", ...}
```

---

## New Data Flow — GET /{code}

```mermaid
sequenceDiagram
    participant Client
    participant Handler as Redirect() (handler/url.go)
    participant Store as SQLiteStore (store/sqlite.go)
    participant DB as urls.db (SQLite file)

    Client->>Handler: GET /xK92pQ
    Handler->>Store: Get("xK92pQ")
    Store->>DB: SELECT * FROM urls WHERE short_code = 'xK92pQ'
    DB-->>Store: row found
    Store-->>Handler: {OriginalURL: "https://youtube.com"}
    Handler-->>Client: 302 Redirect → https://youtube.com
```

---

## Expected Learning

* Persistent storage
* SQL
* SQLite
* CRUD operations
* Database connections
* Repository pattern
* Dependency inversion
