package config

// ProductMongoURIKey selects the product database for CRUD and stock operations.
const ProductMongoURIKey = "PRODUCTS_MONGODB_URI"

// MongoDatabase maps a logical store to the environment key containing its URI.
type MongoDatabase struct {
	Name   string
	URIKey string
}

// MongoDatabases returns the databases opened by the MongoORM adapter.
func MongoDatabases() []MongoDatabase {
	return []MongoDatabase{
		{Name: "products", URIKey: ProductMongoURIKey},
		{Name: "invoices", URIKey: "INVOICES_MONGODB_URI"},
		{Name: "shipping", URIKey: "SHIPPING_MONGODB_URI"},
		{Name: "login", URIKey: "USERS_MONGODB_URI"},
		{Name: "login_sessions", URIKey: "LOGIN_MONGODB_URI"},
	}
}
