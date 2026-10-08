# Invoices service

Invoice snapshots and pending / confirming / complete / rejected lifecycle.

`Repository[Invoice]` persists through MongoORM. `ProductsPort` is wired to the
`productsHTTPAdapter` in `internal/adapter.go`. Confirmation locks the snapshot,
deducts stock with the invoice ID as its idempotency key, then completes the
invoice. Retry confirming after ambiguous upstream failures. Pending invoices
can be edited or rejected.

`POST /api/invoices/auth/token` exchanges a Firebase ID token from the frontend for an
invoice-only JWT. Invoice reads are public. Create, update, confirm and reject
require that JWT and `role: "admin"` on the current Login user record;
the role is loaded on each request. Assign the initial admin role in the
`nx_users.users` collection for the intended account.

## Layout

`main.go` handles bootstrap and shutdown. `internal/controller.go` owns HTTP input and validation; `internal/service.go` owns business rules and calls repository or integration ports. `internal/adapter.go` configures repository endpoints, service URLs and concrete HTTP clients. Controllers call service methods and never access repositories directly.

## Configuration

`INVOICES_PORT`, `PRODUCTS_SERVICE_URL`, `ADAPTERS_SERVICE_URL`, `LOGIN_SERVICE_URL`, `APP_JWT_SECRET`, `INTERNAL_API_KEY`.

Run independently from the workspace root with `go run ./apps/invoices`, or use `make start` for all six processes. See the [root guide](../../README.md) for startup and [API reference](../../docs/api.md) for routes.
