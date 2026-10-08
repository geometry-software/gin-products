# Login service

Firebase ID token verification and public user directory.

The frontend signs in with Firebase and sends its ID token to
`POST /api/auth/session`. Login verifies its signature, project audience, issuer
and timestamps using Firebase public certificates, then provisions or finds a user.
A typed MongoORM user
repository maps each Firebase UID to a stable UUID in the database selected by
`USERS_MONGODB_URI`. Invoices uses the frontend's Firebase ID token to issue its
own JWT. Passwords are never sent to this backend or stored in MongoDB.

## Layout

`main.go` handles bootstrap and shutdown. `internal/controller.go` owns HTTP input and validation; `internal/service.go` owns business rules and calls repository or integration ports. `internal/adapter.go` configures repository endpoints, service URLs and concrete HTTP clients. Controllers call service methods and never access repositories directly.

## Configuration

`LOGIN_PORT`, `ADAPTERS_SERVICE_URL`, `FIREBASE_PROJECT_ID`, `INTERNAL_API_KEY`.

Run independently from the workspace root with `go run ./apps/login`, or use `make start` for all six processes. See the [root guide](../../README.md) for startup and [API reference](../../docs/api.md) for routes.
