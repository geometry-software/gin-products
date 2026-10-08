package invoices

import (
	"context"
	"strings"
	"time"

	"github.com/geometry-software/gin-products/apps/providers/shared/config"
	"github.com/geometry-software/gin-products/apps/providers/shared/controller"
	"github.com/geometry-software/gin-products/apps/providers/shared/http"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const tokenIssuer = "gin-products-invoices"
const tokenAudience = "gin-products-invoices-api"

// IssueToken exchanges a verified Firebase identity for an invoice-only JWT.
func (s *Service) IssueToken(ctx context.Context, idToken string) (TokenResult, error) {
	user, err := s.Users.Exchange(ctx, idToken)
	if err != nil {
		return TokenResult{}, err
	}
	if !user.Active {
		return TokenResult{}, http.Fail(403, "User is inactive")
	}
	expires := time.Now().UTC().Add(time.Hour)
	claims := jwt.RegisteredClaims{
		Issuer:    tokenIssuer,
		Subject:   user.ID,
		Audience:  jwt.ClaimStrings{tokenAudience},
		ExpiresAt: jwt.NewNumericDate(expires),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(config.Get("APP_JWT_SECRET", "")))
	if err != nil {
		return TokenResult{}, err
	}
	return TokenResult{AccessToken: signed, TokenType: "Bearer", ExpiresAt: expires, User: user}, nil
}

func (s *Service) authenticate(c *gin.Context) {
	raw := c.GetHeader("Authorization")
	if !strings.HasPrefix(raw, "Bearer ") {
		controller.Error(c, http.Fail(401, "Bearer token required"))
		return
	}
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(strings.TrimPrefix(raw, "Bearer "), claims,
		func(*jwt.Token) (any, error) { return []byte(config.Get("APP_JWT_SECRET", "")), nil },
		jwt.WithValidMethods([]string{"HS256"}), jwt.WithIssuer(tokenIssuer),
		jwt.WithAudience(tokenAudience), jwt.WithExpirationRequired())
	if err != nil || !token.Valid || claims.Subject == "" {
		controller.Error(c, http.Fail(401, "Invalid or expired token"))
		return
	}
	user, err := s.Users.Get(c.Request.Context(), claims.Subject)
	if err != nil {
		controller.Error(c, http.Fail(401, "User unavailable"))
		return
	}
	if !user.Active {
		controller.Error(c, http.Fail(403, "User is inactive"))
		return
	}
	c.Set("invoiceUser", user)
	c.Next()
}
