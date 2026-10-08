package config

// CollectionKind identifies a typed MongoORM collection owned by a service.
type CollectionKind string

const (
	// ProductsCollection stores catalog products.
	ProductsCollection CollectionKind = "products"
	// InvoicesCollection stores invoices.
	InvoicesCollection CollectionKind = "invoices"
	// ShipmentsCollection stores shipments.
	ShipmentsCollection CollectionKind = "shipments"
	// UsersCollection stores login users.
	UsersCollection CollectionKind = "users"
	// SessionsCollection stores legacy login sessions.
	SessionsCollection CollectionKind = "sessions"
)

// MongoCollection describes an owned collection and its database.
type MongoCollection struct {
	Kind       CollectionKind
	Owner      string
	Database   string
	Collection string
}

// MongoCollections returns the collections exposed by the MongoORM adapter.
func MongoCollections() []MongoCollection {
	return []MongoCollection{
		{Kind: ProductsCollection, Owner: "products", Database: "products", Collection: "products"},
		{Kind: InvoicesCollection, Owner: "invoices", Database: "invoices", Collection: "invoices"},
		{Kind: ShipmentsCollection, Owner: "shipping", Database: "shipping", Collection: "shipments"},
		{Kind: UsersCollection, Owner: "login", Database: "login", Collection: "users"},
		{Kind: SessionsCollection, Owner: "login", Database: "login_sessions", Collection: "sessions"},
	}
}
