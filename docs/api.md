# HTTP API

All requests and responses use JSON except 204 responses. Error shape:
`{"error":"Message"}`. Invalid input returns 400, missing/invalid authentication 401,
inactive users 403, missing records 404, stale writes or business conflicts 409,
and unavailable infrastructure 503.

Login, Products, Shipping and invoice GET routes accept requests without a bearer
token. Invoice writes require an invoice-only JWT and the current user's `admin`
role. Internal routes still require scoped service credentials.
Public services answer browser CORS preflight requests and allow any origin.

## Public routes

| Service | Method | Path | Behavior |
| --- | --- | --- | --- |
| Login | POST | `/api/auth/session` | Verify frontend Firebase ID token and provision user |
| Login | GET | `/api/users` | Paginated user directory |
| Login | GET | `/api/users/:id` | User by ID |
| Products | POST | `/api/products` | Create catalog product; 201 |
| Products | GET | `/api/products` | Paginated catalog |
| Products | GET | `/api/products/:id` | Product by ID |
| Products | PUT | `/api/products/:id` | Replace editable fields; current `version` required |
| Products | DELETE | `/api/products/:id` | Delete product; 204 |
| Invoices | POST | `/api/invoices/auth/token` | Exchange Firebase ID token for invoice JWT |
| Invoices | POST | `/api/invoices` | Resolve products and create pending snapshot; 201 |
| Invoices | GET | `/api/invoices` | Public paginated invoices |
| Invoices | GET | `/api/invoices/:id` | Public invoice by ID |
| Invoices | PUT | `/api/invoices/:id` | Edit pending invoice; current `version` required |
| Invoices | POST | `/api/invoices/:id/confirm` | Idempotent confirmation and stock deduction |
| Invoices | POST | `/api/invoices/:id/reject` | Reject pending invoice |
| Shipping | POST | `/api/shippings` | Snapshot completed invoices; 201 |
| Shipping | GET | `/api/shippings` | Paginated shipments |
| Shipping | GET | `/api/shippings/:id` | Shipment with tracking events |
| Shipping | PATCH | `/api/shippings/:id/status` | Change state; current `version` required |
| All | GET | `/readyz` | MongoDB dependency readiness; unauthenticated |

List endpoints support `page` (1..1000000), `limit` (1..100), `search` (name,
case-insensitive literal), `active` and `status` where applicable. Sort is fixed:
newest `createdAt` first, ID as tie-breaker. Response:
`{ "items": [...], "total": 42, "page": 1, "limit": 20 }`.

## Internal routes

| Owner | Path | Permitted caller |
| --- | --- | --- |
| Adapters | CRUD `/internal/mongoorm/products` | Products |
| Adapters | CRUD `/internal/mongoorm/invoices` | Invoices |
| Adapters | CRUD `/internal/mongoorm/shipments` | Shipping |
| Adapters | CRUD `/internal/mongoorm/users`, `/sessions` | Login |
| Providers | POST `/internal/providers/stock/deduct` | Products |
| Products | POST `/internal/products/resolve` | Invoices |
| Products | POST `/internal/products/stock/deduct` | Invoices |
| Invoices | POST `/internal/invoices/resolve` | Shipping |

Resolve accepts `{ "ids": ["uuid"] }` and returns an array of snapshots. Deduction
accepts `{ "operationId": "invoice UUID", "items": [{ "productId": "uuid",
"quantity": 2 }] }`. The exact item set must stay the same for a repeated operation.
