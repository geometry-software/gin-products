# Adapters service

MongoORM infrastructure adapter and shared service contracts.

`internal/controller/routes.go` registers fixed typed collections and owner credentials. `internal/adapter.go` prepares MongoDB connections and indexes. MongoORMAdapter owns database connections, BSON entity mapping, pagination and version checks. Stock transactions live in Providers service. Generic collection routes only accept their owning service credential. The adapter service is a shared infrastructure dependency, not a business-domain orchestrator.

## Layout

`main.go` loads config and runs the server. Service-specific code lives under `internal/`: `adapter.go` initializes MongoDB and indexes, `http/router.go` creates the router, and `controller/routes.go` registers collections and `/readyz`. The `mongodb/` folder contains the connection adapter, collection controller, repository and indexes. Cross-service HTTP, controller, MongoORM client, configuration and models live in `apps/providers/shared/`; Go's `internal/` import rule keeps service-specific code private.

## Configuration

`ADAPTERS_PORT`, `USERS_MONGODB_URI`, `LOGIN_MONGODB_URI`,
`PRODUCTS_MONGODB_URI`, `INVOICES_MONGODB_URI`, `SHIPPING_MONGODB_URI`,
`INTERNAL_API_KEY`.

Run independently from the workspace root with `go run ./apps/adapters`, or use `make start` for all six processes. See the [root guide](../../README.md) for startup and [API reference](../../docs/api.md) for routes.
