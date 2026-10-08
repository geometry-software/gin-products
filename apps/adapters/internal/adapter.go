package adapters

import (
	"context"

	"github.com/geometry-software/gin-products/apps/adapters/internal/mongodb"
	"github.com/geometry-software/gin-products/apps/providers/shared/http"
)

// New opens MongoDB connections and initializes their indexes.
func New(ctx context.Context) (*mongodb.MongoORMAdapter, error) {
	mongo, err := mongodb.Open(ctx)
	if err != nil {
		return nil, err
	}
	if err := mongo.Indexes(ctx); err != nil {
		mongo.Close(ctx)
		return nil, http.Fail(503, "Cannot initialize MongoDB indexes")
	}
	return mongo, nil
}
