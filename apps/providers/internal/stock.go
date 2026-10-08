package providers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"github.com/geometry-software/gin-products/apps/providers/shared/controller"
	"github.com/geometry-software/gin-products/apps/providers/shared/http"
	"github.com/geometry-software/gin-products/apps/providers/shared/models"
	mongoormclient "github.com/geometry-software/gin-products/apps/providers/shared/mongoorm"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/writeconcern"
)

func registerStock(r *gin.Engine, db *mongo.Database) {
	r.POST("/internal/providers/stock/deduct", http.Internal("products"), func(c *gin.Context) {
		var req models.StockRequest
		if !controller.ValidateJSON(c, &req) {
			return
		}
		if err := DeductStock(c.Request.Context(), db, req); err != nil {
			controller.Error(c, err)
			return
		}
		c.JSON(200, gin.H{"operationId": req.OperationID, "status": "applied"})
	})
}

// DeductStock commits product updates and an idempotency receipt atomically.
// MongoDB must run as a replica set; no standalone fallback is used.
func DeductStock(ctx context.Context, db *mongo.Database, req models.StockRequest) error {
	items, err := mongoormclient.Normalize(req.Items)
	if err != nil {
		return err
	}
	data, _ := json.Marshal(items)
	hash := sha256.Sum256(data)
	digest := hex.EncodeToString(hash[:])
	session, err := db.Client().StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)
	_, err = session.WithTransaction(ctx, func(tx context.Context) (any, error) {
		var receipt struct {
			Digest string `bson:"digest"`
		}
		e := db.Collection("stock_operations").FindOne(tx, bson.M{"_id": req.OperationID}).Decode(&receipt)
		if e == nil {
			if receipt.Digest != digest {
				return nil, http.Fail(409, "Operation ID reused with different items")
			}
			return nil, nil
		}
		if !errors.Is(e, mongo.ErrNoDocuments) {
			return nil, e
		}
		for _, item := range items {
			res, e := db.Collection("products").UpdateOne(tx, bson.M{"_id": item.ProductID, "active": true, "quantity": bson.M{"$gte": item.Quantity}}, bson.M{"$inc": bson.M{"quantity": -item.Quantity, "version": 1}, "$set": bson.M{"updatedAt": time.Now().UTC()}})
			if e != nil {
				return nil, e
			}
			if res.MatchedCount != 1 {
				return nil, http.Fail(409, "Product unavailable or insufficient stock")
			}
		}
		_, e = db.Collection("stock_operations").InsertOne(tx, bson.M{"_id": req.OperationID, "digest": digest, "createdAt": time.Now().UTC()})
		return nil, e
	}, options.Transaction().SetWriteConcern(writeconcern.Majority()))
	if mongo.IsDuplicateKeyError(err) {
		return http.Fail(409, "Record already exists")
	}
	return err
}
