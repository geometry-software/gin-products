package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Indexes ensures the required user and session indexes exist.
func (adapter *MongoORMAdapter) Indexes(ctx context.Context) error {
	_, err := adapter.databases["login"].Collection("users").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "firebaseUid", Value: 1}},
		Options: options.Index().SetUnique(true).SetPartialFilterExpression(
			bson.M{"firebaseUid": bson.M{"$type": "string"}},
		),
	})
	if err != nil {
		return err
	}
	_, err = adapter.databases["login_sessions"].Collection("sessions").Indexes().CreateOne(ctx,
		mongo.IndexModel{Keys: bson.D{{Key: "expiresAt", Value: 1}}, Options: options.Index().SetExpireAfterSeconds(0)})
	return err
}
