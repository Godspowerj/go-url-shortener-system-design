# Workspace Rules for URL Shortener System Design

## Architecture Workflow Rules

All developers and agentic assistants modifying this codebase MUST follow this sequence before implementing any major architectural changes (e.g., adding cache, switching database, adding background queues):

1. **Identify Problem:** Clearly write down the bottleneck or limitation (e.g., reads are slow, server crashes under write load).
2. **Benchmark & Measure:** Measure performance before making changes. Save the numbers.
3. **Write ADR:** Create an Architecture Decision Record under `docs/adrs/` explaining the chosen path, trade-offs, and alternatives considered.
4. **Draw Architecture:** Update or add a diagram under `docs/architecture/`.
5. **Implement:** Write and refactor the code.
6. **Benchmark Again:** Run the same load test to measure the difference.
7. **Document Results:** Record the new numbers under `docs/benchmarks/`.

---

## ADR Template

Every ADR must follow this format:

```markdown
# ADR-[Number]: [Title]

- **Status:** [Proposed | Accepted | Superseded | Deprecated]
- **Date:** YYYY-MM-DD

## Context
What problem are we facing? Why is the current architecture insufficient?

## Decision
What is the chosen solution?

## Alternatives Considered
What else did we consider, and why did we reject it?

## Pros and Cons
- **Pros:** ...
- **Cons:** ...

## Consequences
What are the long-term trade-offs or technical debt we are accepting?

## Validation
How did we or will we validate this decision (e.g., load tests, benchmarks, metrics)?
```
