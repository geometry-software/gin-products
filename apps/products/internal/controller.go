package products

import (
	"github.com/geometry-software/gin-products/apps/providers/shared/controller"
	"github.com/geometry-software/gin-products/apps/providers/shared/http"
	"github.com/geometry-software/gin-products/apps/providers/shared/models"
	"github.com/gin-gonic/gin"
)

// Routes registers public product endpoints and scoped service-to-service routes.
func Routes(r *gin.Engine, s *Service) {
	g := r.Group("/api/products")
	g.POST("", func(c *gin.Context) {
		product, ok := controller.BindJSON[models.Product](c)
		if !ok {
			return
		}
		controller.JSON(c, 201, func() (models.Product, error) { return s.Create(c.Request.Context(), product) })
	})
	g.GET("", func(c *gin.Context) {
		controller.JSON(c, 200, func() (models.Page[models.Product], error) { return s.List(c.Request.Context(), c.Request.URL.Query()) })
	})
	g.GET("/:id", func(c *gin.Context) {
		controller.JSON(c, 200, func() (models.Product, error) { return s.Get(c.Request.Context(), c.Param("id")) })
	})
	g.PUT("/:id", func(c *gin.Context) {
		product, ok := controller.BindJSON[models.Product](c)
		if !ok {
			return
		}
		controller.JSON(c, 200, func() (models.Product, error) { return s.Update(c.Request.Context(), c.Param("id"), product) })
	})
	g.DELETE("/:id", func(c *gin.Context) {
		controller.NoContent(c, func() error { return s.Delete(c.Request.Context(), c.Param("id")) })
	})
	internal := r.Group("/internal/products", http.Internal("invoices"))
	internal.POST("/resolve", func(c *gin.Context) {
		input, ok := controller.BindJSON[struct {
			IDs []string `json:"ids" binding:"required,min=1,max=100,dive,uuid"`
		}](c)
		if !ok {
			return
		}
		controller.JSON(c, 200, func() ([]models.Product, error) { return s.Resolve(c.Request.Context(), input.IDs) })
	})
	internal.POST("/stock/deduct", func(c *gin.Context) {
		input, ok := controller.BindJSON[models.StockRequest](c)
		if !ok {
			return
		}
		if err := s.Deduct(c.Request.Context(), input); err != nil {
			controller.Error(c, err)
			return
		}
		c.JSON(200, gin.H{"status": "applied", "operationId": input.OperationID})
	})
}
