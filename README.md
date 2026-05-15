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

# Architectural Solutions
To ensure high performance and reduce overhead on the Go runtime, the project utilizes Nginx as a high-performance reverse proxy and static file server.

Static Assets: Swagger UI (HTML/JS/CSS) and the swagger.json specification are served directly by Nginx from the filesystem.

API Traffic: All requests to /v1/* are proxied to the Go backend service.

Database Initialization: The system includes an automated SQL seed that populates a default user and wallet for immediate testing upon deployment.

### Why is the wallet a separate entity?
I decided against storing balances directly in the user table (flat structure) when using a 1:1 relationship (User -> Wallet) for the following reasons:

- Domain Isolation:
  Users are responsible for identification and profiles (login, email, hash table). The wallet table is responsible for financial status. Changing the profile logic should not affect financial transactions, and vice versa.

- Locking and Concurrency (Lock Strategy):
  We use SELECT FOR UPDATE when synchronizing.

- Scalability and Sharding:
  In the future, financial data may grow faster than profiles. Separate tables allow wallets to be moved to a separate disk or even to specific database points (sharding) without having to redesign the entire authorization system.

- Auditing and Security:
  Financial tables often require stricter security policies (row-level security) and detailed logging (history triggers). It's easier to set this up on a compact wallet table than on a small user table with a lot of metadata.
