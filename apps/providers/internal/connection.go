package providers

import (
	"context"
	"net/url"
	"strings"
	"time"

	"github.com/geometry-software/gin-products/apps/providers/shared/http"
	"github.com/geometry-software/gin-products/apps/providers/shared/config"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

// Connection holds the stock provider's MongoDB connection.
type Connection struct {
	Client   *mongo.Client
	Database *mongo.Database
}

// Open connects to the product database using PRODUCTS_MONGODB_URI.
func Open(ctx context.Context) (*Connection, error) {
	uri := config.Get(config.ProductMongoURIKey, "")
	u, err := url.Parse(uri)
	if err != nil || u == nil || (u.Scheme != "mongodb" && u.Scheme != "mongodb+srv") || strings.Trim(u.Path, "/") == "" {
		return nil, http.Fail(500, "Missing or invalid "+config.ProductMongoURIKey+" with database name")
	}
	client, err := mongo.Connect(options.Client().ApplyURI(uri).
		SetBSONOptions(&options.BSONOptions{ObjectIDAsHexString: true}).
		SetServerSelectionTimeout(5 * time.Second))
	if err != nil {
		return nil, http.Fail(503, "Cannot configure MongoDB for stock")
	}
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		_ = client.Disconnect(ctx)
		return nil, http.Fail(503, "Cannot connect to MongoDB for stock")
	}
	return &Connection{Client: client, Database: client.Database(strings.Trim(u.Path, "/"))}, nil
}

// Ready checks the stock provider's MongoDB connection.
func (c *Connection) Ready(ctx context.Context) error {
	if err := c.Client.Ping(ctx, readpref.Primary()); err != nil {
		return http.Fail(503, "MongoDB unavailable")
	}
	return nil
}

// Close disconnects the stock provider's MongoDB client.
func (c *Connection) Close(ctx context.Context) { _ = c.Client.Disconnect(ctx) }
