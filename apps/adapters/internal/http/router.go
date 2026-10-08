package http

import (
	"github.com/geometry-software/gin-products/apps/adapters/internal/controller"
	"github.com/geometry-software/gin-products/apps/adapters/internal/mongodb"
	sharedhttp "github.com/geometry-software/gin-products/apps/providers/shared/http"
	"github.com/gin-gonic/gin"
)

// Router creates the Adapters HTTP router and registers its routes.
func Router(adapter *mongodb.MongoORMAdapter) *gin.Engine {
	router := sharedhttp.New("adapters")
	controller.Routes(router, adapter)
	return router
}
