# v3 Architecture: SQLite Persistence with Redis Caching

This diagram illustrates how read queries are cached in Redis to avoid hitting the SQLite database for every redirect request.

```mermaid
graph TD
    Client["Client (Browser / curl)"] -->|1. GET /:code| Router["Chi Router"]
    Router -->|2. Redirect()| Handler["handler/url.go"]
    Handler -->|3. ResolveURL(:code)| Service["service/service.go"]
    
    Service -->|4. Get(:code)| Redis["store/redis.go (Redis Cache)"]
    
    Redis -.->|5a. Cache Hit (Fast path)| Service
    
    Service -->|5b. Cache Miss (Slow path)| SQLite["store/sqlite.go (SQLite Database)"]
    SQLite -.->|6. URL details| Service
    Service -->|7. Set(:code, URL)| Redis
    
    Service -->|8. Push ID to clicksQueue| Queue["clicksQueue (Go Channel)"]
    Queue -->|9. Process Click Asynchronously| Worker["Worker Goroutine"]
    Worker -->|10. IncrementClick(id)| SQLite
    
    Service -->|11. Return Original URL| Handler
    Handler -->|12. 302 Redirect| Client
```
