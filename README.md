# Prelude
This project is built with an infrastructure-first mindset. I deliberately used commented conventions to highlight areas for extending functionality. While this version is simplified and focused on core endpoints to respect review time, extreme extensibility, high concurrency support, and fault tolerance were built in from the ground up.

### Key Features
- Fin-Tech Grade Reliability: Redis-based idempotency layer to prevent double-spending on network retries.

- Atomic Operations: Unit of Work pattern (TxManager) for strict ACID compliance and business logic isolation.

- High-Performance Tracing: Fast non-cryptographic hashing (FNV-1a) for public transaction IDs (TXID), completely obfuscating internal database sequences.

- Edge Performance: Nginx acting as a reverse proxy and static asset server to offload the Go runtime.

- Observability: Structured JSON logging for seamless integration with ELK/Grafana stacks.

# Quick Start

Follow these steps to spin up the entire infrastructure (Postgres + Go Backend + Nginx):
```bash
make set-env
make build
make start
make migrate
```

| Resource              | Path                                         | Description                            |
|:----------------------|:---------------------------------------------|:---------------------------------------|
| **Swagger UI**        | `http://localhost:8080/swagger/`             | Interactive API documentation (Static) |
| **API Specification** | `http://localhost:8080/swagger/swagger.json` | Raw OpenAPI 2.0 / 3.0 spec             |


Note: The system includes an automated SQL seed that populates a default user and wallet for immediate testing upon deployment.

# Architectural Solutions
1. Edge Layer & Static Offloading
   To ensure maximum throughput and reduce garbage collection overhead on the Go runtime, Nginx is utilized as a high-performance reverse proxy.

   - Static Assets: Swagger UI (HTML/JS/CSS) and the swagger.json specification are served directly from the filesystem by Nginx.

   - API Traffic: All dynamic requests to /v1/* are proxied to the Go backend.

2. Idempotency & Safe Retries
   In financial systems, network timeouts are inevitable. The API implements a strict idempotency layer using Redis.

   - Mechanics: Mutating requests (like payouts) require an X-Idempotency-Key. The middleware strictly locks the request to prevent race conditions and caches the full HTTP response (Status + Body).

   - Benefit: Clients can safely retry requests upon connection drops without the risk of triggering duplicate database transactions.

3. Domain Isolation: Why is the Wallet a separate entity?
   I decided against storing balances directly in the user table (flat structure) when using a 1:1 relationship (User -> Wallet) for several critical reasons:

   - Domain Segregation: Users are responsible for identification (login, auth). The wallet table is strictly responsible for financial state. Modifying profile logic will never inadvertently impact financial transactions.

   - Locking Strategy: Financial mutations require pessimistic locks (SELECT ... FOR UPDATE). Separating the wallet allows us to lock only the balance row, leaving the user profile fully accessible for parallel read operations.

   - Scalability & Sharding: Financial audit logs and balances grow exponentially faster than user profiles. This separation allows wallets to be moved to specific fast-storage tiers or entirely different shards without redesigning the authorization domain.

4. Transaction Management (Unit of Work)
   Business logic is completely decoupled from database state management.

   - The Service layer orchestrates business rules (e.g., insufficient funds validation) and delegates state changes to "dumb" repositories.

   - A custom TxManager propagates pgx.Tx through context.Context, guaranteeing that all balance updates and audit log insertions either commit atomically or roll back cleanly if any business constraint fails.
