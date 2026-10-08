package mongodb

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

// MongoORMAdapter holds MongoDB clients and their selected databases.
type MongoORMAdapter struct {
	clients   []*mongo.Client
	databases map[string]*mongo.Database
}

// Open connects to the configured MongoDB databases.
func Open(ctx context.Context) (*MongoORMAdapter, error) {
	adapter := &MongoORMAdapter{databases: map[string]*mongo.Database{}}
	for _, database := range config.MongoDatabases() {
		key := database.URIKey
		uri := config.Get(key, "")
		u, err := url.Parse(uri)
		if err != nil || (u.Scheme != "mongodb" && u.Scheme != "mongodb+srv") || strings.Trim(u.Path, "/") == "" {
			adapter.Close(ctx)
			return nil, http.Fail(500, "Missing or invalid "+key+" with database name")
		}
		client, err := mongo.Connect(options.Client().ApplyURI(uri).
			SetBSONOptions(&options.BSONOptions{ObjectIDAsHexString: true}).
			SetServerSelectionTimeout(5 * time.Second))
		if err != nil {
			adapter.Close(ctx)
			return nil, http.Fail(503, "Cannot configure MongoDB for "+database.Name)
		}
		adapter.clients = append(adapter.clients, client)
		if err = client.Ping(ctx, readpref.Primary()); err != nil {
			adapter.Close(ctx)
			return nil, http.Fail(503, "Cannot connect to MongoDB for "+database.Name)
		}
		adapter.databases[database.Name] = client.Database(strings.Trim(u.Path, "/"))
	}
	return adapter, nil
}

// Close disconnects every MongoDB client.
func (adapter *MongoORMAdapter) Close(ctx context.Context) {
	for _, c := range adapter.clients {
		_ = c.Disconnect(ctx)
	}
}

// Ready checks that every MongoDB connection is available.
func (adapter *MongoORMAdapter) Ready(ctx context.Context) error {
	for _, c := range adapter.clients {
		if err := c.Ping(ctx, readpref.Primary()); err != nil {
			return http.Fail(503, "MongoDB unavailable")
		}
	}
	return nil
}
