package mongodb

import (
	"context"
	"errors"
	"time"

	"github.com/geometry-software/gin-products/apps/providers/shared/http"
	"github.com/geometry-software/gin-products/apps/providers/shared/models"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type collectionRepository[T any] struct{ collection *mongo.Collection }

func (r collectionRepository[T]) create(ctx context.Context, value T) (T, error) {
	metadata := meta(&value)
	if metadata.ID == "" {
		metadata.ID = uuid.NewString()
	} else if uuid.Validate(metadata.ID) != nil {
		return value, http.Fail(400, "Invalid ID")
	}
	metadata.Version = 1
	metadata.CreatedAt = time.Now().UTC()
	metadata.UpdatedAt = metadata.CreatedAt
	if _, err := r.collection.InsertOne(ctx, value); err != nil {
		return value, Translate(err)
	}
	return value, nil
}

func (r collectionRepository[T]) list(ctx context.Context, filter bson.M, page, limit int64) (models.Page[T], error) {
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return models.Page[T]{}, err
	}
	cursor, err := r.collection.Find(ctx, filter, options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}, {Key: "_id", Value: 1}}).
		SetSkip((page-1)*limit).SetLimit(limit))
	if err != nil {
		return models.Page[T]{}, err
	}
	defer cursor.Close(ctx)
	values := []T{}
	if err := cursor.All(ctx, &values); err != nil {
		return models.Page[T]{}, err
	}
	return models.Page[T]{Items: values, Total: total, Page: page, Limit: limit}, nil
}

func (r collectionRepository[T]) get(ctx context.Context, id string) (T, error) {
	var value T
	if err := r.collection.FindOne(ctx, idFilter(id)).Decode(&value); err != nil {
		return value, Translate(err)
	}
	return value, nil
}

func (r collectionRepository[T]) replace(ctx context.Context, id string, value T) (T, error) {
	metadata := meta(&value)
	if metadata.ID != id || metadata.Version < 1 {
		return value, http.Fail(400, "ID and version are required")
	}
	version := metadata.Version
	metadata.Version++
	metadata.UpdatedAt = time.Now().UTC()
	result, err := r.collection.ReplaceOne(ctx, bson.M{"_id": metadata.ID, "version": version}, value)
	if err != nil {
		return value, Translate(err)
	}
	if result.MatchedCount == 0 {
		return value, http.Fail(409, "Record changed; reload and retry")
	}
	return value, nil
}

func (r collectionRepository[T]) delete(ctx context.Context, id string, version int64) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id, "version": version})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return http.Fail(409, "Record changed or missing")
	}
	return nil
}

// Translate maps MongoDB errors to HTTP domain errors.
func Translate(err error) error {
	if errors.Is(err, mongo.ErrNoDocuments) {
		return http.Fail(404, "Record not found")
	}
	if mongo.IsDuplicateKeyError(err) {
		return http.Fail(409, "Record already exists")
	}
	return err
}

func meta[T any](value *T) *models.Entity {
	return any(value).(interface{ Meta() *models.Entity }).Meta()
}

func idFilter(id string) bson.M {
	if objectID, err := bson.ObjectIDFromHex(id); err == nil {
		return bson.M{"_id": bson.M{"$in": bson.A{id, objectID}}}
	}
	return bson.M{"_id": id}
}
