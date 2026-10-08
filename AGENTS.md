# Repository instructions for agents

Read the [root README](README.md) before changing a service. It is the canonical
guide to the service map, Go directory layout, local startup and HTTP contracts.
Keep the project backend-only and keep credentials out of committed files.

## Creating a microservice

Follow [Create a microservice](README.md#create-a-microservice) in order. Define
ownership and HTTP dependencies first, then create `main.go`, `controller.go`,
`service.go`, `adapter.go` and the service README. Keep HTTP input and responses in
the controller, business rules in the service, and connection configuration plus
concrete HTTP/repository clients in the adapter. Business services use MongoORM
through the Adapters process; they never connect to MongoDB directly. Login
verifies frontend-supplied Firebase ID tokens against Firebase public signing keys.
Use the shared `controller` for JSON validation and HTTP error responses.
Only Invoices uses bearer JWTs for writes; reads are public and its controller
checks the `admin` role on writes.
Document exported Go APIs with GoDoc comments and explicit result types.
