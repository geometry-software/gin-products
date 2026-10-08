package invoices

import (
	"github.com/geometry-software/gin-products/apps/providers/shared/controller"
	"github.com/geometry-software/gin-products/apps/providers/shared/http"
	"github.com/geometry-software/gin-products/apps/providers/shared/models"
	"github.com/gin-gonic/gin"
)

func requireAdmin(c *gin.Context) {
	value, found := c.Get("invoiceUser")
	user, ok := value.(models.User)
	if !found || !ok || user.Role != "admin" {
		controller.Error(c, http.Fail(403, "Admin role required"))
		return
	}
	c.Next()
}

// Routes registers public invoice reads, admin writes, and internal reads.
func Routes(r *gin.Engine, s *Service) {
	r.POST("/api/invoices/auth/token", func(c *gin.Context) {
		input, ok := controller.BindJSON[struct {
			IDToken string `json:"idToken" binding:"required,max=16384"`
		}](c)
		if !ok {
			return
		}
		controller.JSON(c, 200, func() (TokenResult, error) { return s.IssueToken(c.Request.Context(), input.IDToken) })
	})
	g := r.Group("/api/invoices")
	g.GET("", func(c *gin.Context) {
		controller.JSON(c, 200, func() (models.Page[models.Invoice], error) { return s.List(c.Request.Context(), c.Request.URL.Query()) })
	})
	g.GET("/:id", func(c *gin.Context) {
		controller.JSON(c, 200, func() (models.Invoice, error) { return s.Get(c.Request.Context(), c.Param("id")) })
	})
	writes := g.Group("", s.authenticate, requireAdmin)
	writes.POST("", func(c *gin.Context) {
		input, ok := controller.BindJSON[Input](c)
		if !ok {
			return
		}
		controller.JSON(c, 201, func() (models.Invoice, error) { return s.Create(c.Request.Context(), input) })
	})
	writes.PUT("/:id", func(c *gin.Context) {
		input, ok := controller.BindJSON[Input](c)
		if !ok {
			return
		}
		controller.JSON(c, 200, func() (models.Invoice, error) { return s.Update(c.Request.Context(), c.Param("id"), input) })
	})
	writes.POST("/:id/confirm", func(c *gin.Context) {
		controller.JSON(c, 200, func() (models.Invoice, error) { return s.Confirm(c.Request.Context(), c.Param("id")) })
	})
	writes.POST("/:id/reject", func(c *gin.Context) {
		controller.JSON(c, 200, func() (models.Invoice, error) { return s.Reject(c.Request.Context(), c.Param("id")) })
	})
	r.POST("/internal/invoices/resolve", http.Internal("shipping"), func(c *gin.Context) {
		input, ok := controller.BindJSON[struct {
			IDs []string `json:"ids" binding:"required,min=1,max=100,dive,uuid"`
		}](c)
		if !ok {
			return
		}
		controller.JSON(c, 200, func() ([]models.Invoice, error) { return s.Resolve(c.Request.Context(), input.IDs) })
	})
}
