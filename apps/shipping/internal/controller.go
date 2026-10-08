package shipping

import (
	"github.com/geometry-software/gin-products/apps/providers/shared/controller"
	"github.com/geometry-software/gin-products/apps/providers/shared/models"
	"github.com/gin-gonic/gin"
)

// Routes registers the public shipment API.
func Routes(r *gin.Engine, s *Service) {
	g := r.Group("/api/shippings")
	g.POST("", func(c *gin.Context) {
		input, ok := controller.BindJSON[Input](c)
		if !ok {
			return
		}
		controller.JSON(c, 201, func() (models.Shipment, error) { return s.Create(c.Request.Context(), input) })
	})
	g.GET("", func(c *gin.Context) {
		controller.JSON(c, 200, func() (models.Page[models.Shipment], error) {
			return s.List(c.Request.Context(), c.Request.URL.Query())
		})
	})
	g.GET("/:id", func(c *gin.Context) {
		controller.JSON(c, 200, func() (models.Shipment, error) { return s.Get(c.Request.Context(), c.Param("id")) })
	})
	g.PATCH("/:id/status", func(c *gin.Context) {
		input, ok := controller.BindJSON[struct {
			Status  string `json:"status" binding:"required,oneof=in_transit out_for_delivery delivered exception cancelled"`
			Version int64  `json:"version" binding:"gte=1"`
		}](c)
		if !ok {
			return
		}
		controller.JSON(c, 200, func() (models.Shipment, error) {
			return s.ChangeStatus(c.Request.Context(), c.Param("id"), input.Status, input.Version)
		})
	})
}
