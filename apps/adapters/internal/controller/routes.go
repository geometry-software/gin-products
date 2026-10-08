package controller

import (
	"context"
	"time"

	"github.com/geometry-software/gin-products/apps/adapters/internal/mongodb"
	"github.com/geometry-software/gin-products/apps/providers/shared/config"
	response "github.com/geometry-software/gin-products/apps/providers/shared/controller"
	"github.com/geometry-software/gin-products/apps/providers/shared/models"
	"github.com/gin-gonic/gin"
)

// Routes registers owned collection endpoints and the readiness endpoint.
func Routes(router *gin.Engine, adapter *mongodb.MongoORMAdapter) {
	for _, collection := range config.MongoCollections() {
		switch collection.Kind {
		case config.ProductsCollection:
			mongodb.RegisterInDatabase[models.Product](router, adapter, collection.Owner, collection.Database, collection.Collection)
		case config.InvoicesCollection:
			mongodb.RegisterInDatabase[models.Invoice](router, adapter, collection.Owner, collection.Database, collection.Collection)
		case config.ShipmentsCollection:
			mongodb.RegisterInDatabase[models.Shipment](router, adapter, collection.Owner, collection.Database, collection.Collection)
		case config.UsersCollection:
			mongodb.RegisterInDatabase[models.User](router, adapter, collection.Owner, collection.Database, collection.Collection)
		case config.SessionsCollection:
			mongodb.RegisterInDatabase[models.Session](router, adapter, collection.Owner, collection.Database, collection.Collection)
		}
	}
	router.GET("/readyz", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()
		if err := adapter.Ready(ctx); err != nil {
			response.Error(c, err)
			return
		}
		c.JSON(200, gin.H{"status": "ready"})
	})
}
