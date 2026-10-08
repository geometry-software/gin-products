# Shipping service

Shipments, invoice snapshots and local tracking events.

Repository[Shipment] persists through MongoORM. InvoicesPort uses an HTTP client
to fetch completed invoices. Status changes append local events and use optimistic
version checks. No external delivery/geography providers are implemented.

## Layout

`main.go` handles bootstrap and shutdown. `internal/controller.go` owns HTTP input and validation; `internal/service.go` owns business rules and calls repository or integration ports. `internal/adapter.go` configures repository endpoints, service URLs and concrete HTTP clients. Controllers call service methods and never access repositories directly.

## Configuration

`SHIPPING_PORT`, `INVOICES_SERVICE_URL`, `ADAPTERS_SERVICE_URL`, `INTERNAL_API_KEY`.
Public shipment routes do not require a bearer token, so new shipments have no
authenticated creator.

Run independently from the workspace root with `go run ./apps/shipping`, or use `make start` for all six processes. See the [root guide](../../README.md) for startup and [API reference](../../docs/api.md) for routes.
