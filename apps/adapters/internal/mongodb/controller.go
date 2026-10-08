package mongodb

import (
	"regexp"
	"strconv"

	"github.com/geometry-software/gin-products/apps/providers/shared/controller"
	"github.com/geometry-software/gin-products/apps/providers/shared/http"
	"github.com/geometry-software/gin-products/apps/providers/shared/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// Register maps a concrete domain type to a fixed collection. Clients cannot submit
// MongoDB query operators, database names, update operators or arbitrary fields.
func Register[T any](router *gin.Engine, adapter *MongoORMAdapter, owner, collection string) {
	RegisterInDatabase[T](router, adapter, owner, owner, collection)
}

// RegisterInDatabase keeps the HTTP credential scoped to owner while selecting
// the database that holds this collection. Login owns users and sessions in
// separate databases, but both routes accept only Login's service credential.
func RegisterInDatabase[T any](router *gin.Engine, adapter *MongoORMAdapter, owner, database, collection string) {
	repository := collectionRepository[T]{collection: adapter.databases[database].Collection(collection)}
	group := router.Group("/internal/mongoorm/"+collection, http.Internal(owner))
	group.POST("", func(c *gin.Context) {
		value, ok := controller.BindJSON[T](c)
		if !ok {
			return
		}
		controller.JSON(c, 201, func() (T, error) { return repository.create(c.Request.Context(), value) })
	})
	group.GET("", func(c *gin.Context) {
		page, limit, err := http.Paging(c)
		if err != nil {
			controller.Error(c, err)
			return
		}
		filter, err := listFilter(c)
		if err != nil {
			controller.Error(c, err)
			return
		}
		controller.JSON(c, 200, func() (models.Page[T], error) {
			return repository.list(c.Request.Context(), filter, page, limit)
		})
	})
	group.GET("/:id", func(c *gin.Context) {
		controller.JSON(c, 200, func() (T, error) {
			return repository.get(c.Request.Context(), c.Param("id"))
		})
	})
	group.PUT("/:id", func(c *gin.Context) {
		value, ok := controller.BindJSON[T](c)
		if !ok {
			return
		}
		controller.JSON(c, 200, func() (T, error) {
			return repository.replace(c.Request.Context(), c.Param("id"), value)
		})
	})
	group.DELETE("/:id", func(c *gin.Context) {
		input, ok := controller.BindJSON[struct {
			Version int64 `json:"version" binding:"gte=1"`
		}](c)
		if !ok {
			return
		}
		controller.NoContent(c, func() error {
			return repository.delete(c.Request.Context(), c.Param("id"), input.Version)
		})
	})
}

func listFilter(c *gin.Context) (bson.M, error) {
	filter := bson.M{}
	for _, field := range []string{"status", "firebaseUid", "userId", "email"} {
		if value := c.Query(field); value != "" {
			if field == "email" {
				if len(value) > 320 {
					return nil, http.Fail(400, "Email filter too long")
				}
				filter[field] = bson.M{"$regex": "^" + regexp.QuoteMeta(value) + "$", "$options": "i"}
				continue
			}
			filter[field] = value
		}
	}
	if value := c.Query("active"); value != "" {
		active, err := strconv.ParseBool(value)
		if err != nil {
			return nil, http.Fail(400, "Invalid active filter")
		}
		filter["active"] = active
	}
	if value := c.Query("search"); value != "" {
		if len(value) > 200 {
			return nil, http.Fail(400, "Search too long")
		}
		filter["name"] = bson.M{"$regex": regexp.QuoteMeta(value), "$options": "i"}
	}
	return filter, nil
}
