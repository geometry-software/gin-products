# Products service

Catalog CRUD, product resolution and atomic stock deduction.

The service uses Repository[Product] through RemoteRepository. Invoices resolves catalog snapshots through internal HTTP routes. Stock deductions go to the stock provider in Providers service, which atomically updates stock and its idempotency receipt. Product prices are integer minor units; PUT requires the current version.

## Layout

`main.go` handles bootstrap and shutdown. `internal/controller.go` owns HTTP input and validation; `internal/service.go` owns business rules and calls repository or integration ports. `internal/adapter.go` configures repository endpoints, service URLs and concrete HTTP clients. Controllers call service methods and never access repositories directly.

## Configuration

`PRODUCTS_PORT`, `ADAPTERS_SERVICE_URL`, `PROVIDERS_SERVICE_URL`, `INTERNAL_API_KEY`. Public product routes
do not require a bearer token.

Run independently from the workspace root with `go run ./apps/products`, or use `make start` for all six processes. See the [root guide](../../README.md) for startup and [API reference](../../docs/api.md) for routes.
