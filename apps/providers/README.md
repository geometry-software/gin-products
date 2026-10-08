# Providers service

Providers service hosts named infrastructure operations beyond generic MongoORM
CRUD. `internal/providers.go` contains the provider enum and registry. Each
provider has its own implementation file; `internal/stock.go` is the first.
`shared/config/services.go` contains the cross-service names, default ports and
URL lookup used by all Go services; dotenv and internal credentials are in
`shared/config/config.go`.
`shared/config/databases.go` maps logical MongoDB stores to `.env` URI keys for
the MongoORM adapter and stock connection.
Shared entity and HTTP contract types live in `shared/models/models.go`.

Stock accepts `POST /internal/providers/stock/deduct` from Products using its
internal service credential. It updates quantities and writes an idempotency
receipt in one MongoDB transaction, requiring a replica set. `internal/connection.go`
opens the products database from `PRODUCTS_MONGODB_URI`; `/readyz` checks it.

Configure `PROVIDERS_PORT` (default 3007), `PRODUCTS_MONGODB_URI`, and
`INTERNAL_API_KEY`. Products calls it using `PROVIDERS_SERVICE_URL`. Run
`make start` from the workspace root to start all six services.
