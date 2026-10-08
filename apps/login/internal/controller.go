package login

import (
	"github.com/geometry-software/gin-products/apps/providers/shared/controller"
	"github.com/geometry-software/gin-products/apps/providers/shared/models"
	"github.com/gin-gonic/gin"
)

// Routes registers frontend-token exchange and public user-directory endpoints.
func Routes(r *gin.Engine, s *Service) {
	r.POST("/api/auth/session", func(c *gin.Context) {
		input, ok := controller.BindJSON[struct {
			IDToken string `json:"idToken" binding:"required,max=16384"`
			Name    string `json:"name" binding:"max=200"`
		}](c)
		if !ok {
			return
		}
		controller.JSON(c, 200, func() (models.User, error) { return s.Exchange(c.Request.Context(), input.IDToken, input.Name) })
	})
	r.GET("/api/users", func(c *gin.Context) {
		controller.JSON(c, 200, func() (models.Page[models.User], error) { return s.List(c.Request.Context(), c.Request.URL.Query()) })
	})
	r.GET("/api/users/:id", func(c *gin.Context) {
		controller.JSON(c, 200, func() (models.User, error) { return s.User(c.Request.Context(), c.Param("id")) })
	})
}
