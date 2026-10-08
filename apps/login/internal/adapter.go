package login

import (
	"context"

	"github.com/geometry-software/gin-products/apps/providers/shared/config"
	"github.com/geometry-software/gin-products/apps/providers/shared/models"
	"github.com/geometry-software/gin-products/apps/providers/shared/mongoorm"
)

// Identity is a verified Firebase ID token's account information.
type Identity struct {
	UID           string
	Email         string
	Name          string
	EmailVerified bool
}

// IdentityPort verifies a token supplied by the frontend.
type IdentityPort interface {
	Verify(context.Context, string) (Identity, error)
}

// New wires the user repository and Firebase ID token verifier.
func New() *Service {
	baseURL := config.URL("adapters")
	return &Service{
		Users:    mongoorm.RemoteRepository[models.User]{BaseURL: baseURL, Caller: "login", Collection: "users"},
		Identity: NewTokenVerifier(config.Get("FIREBASE_PROJECT_ID", "")),
	}
}
