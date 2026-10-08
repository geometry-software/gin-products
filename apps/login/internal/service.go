package login

import (
	"context"
	"errors"
	"github.com/geometry-software/gin-products/apps/providers/shared/http"
	"github.com/geometry-software/gin-products/apps/providers/shared/models"
	"github.com/geometry-software/gin-products/apps/providers/shared/mongoorm"
	"github.com/google/uuid"
	"net/url"
	"strings"
)

// Service holds the repositories and integrations for this domain.
type Service struct {
	Users    mongoorm.Repository[models.User]
	Identity IdentityPort
}

// Exchange verifies a Firebase token and provisions its user.
func (s *Service) Exchange(ctx context.Context, idToken, name string) (models.User, error) {
	identity, err := s.Identity.Verify(ctx, idToken)
	if err != nil {
		return models.User{}, err
	}
	// Deterministic IDs make repeated Firebase sign-ins and concurrent creation safe.
	id := uuid.NewSHA1(uuid.NameSpaceURL, []byte("firebase:"+identity.UID)).String()
	user, err := s.Users.Get(ctx, id)
	if err != nil {
		var e *http.Error
		if !errors.As(err, &e) || e.Status != 404 {
			return models.User{}, err
		}
		if identity.Email != "" && identity.EmailVerified {
			matches, listErr := s.Users.List(ctx, url.Values{"email": {strings.ToLower(identity.Email)}, "limit": {"2"}})
			if listErr != nil {
				return models.User{}, listErr
			}
			if matches.Total > 1 {
				return models.User{}, http.Fail(409, "Multiple users share this email")
			}
			if len(matches.Items) == 1 {
				user = matches.Items[0]
				err = nil
			}
		}
		if err == nil {
			if !user.Active {
				return models.User{}, http.Fail(403, "User is inactive")
			}
			return user, nil
		}
		if name == "" {
			name = identity.Name
		}
		if name == "" {
			name = "User"
		}
		user, err = s.Users.Create(ctx, models.User{Entity: models.Entity{ID: id}, FirebaseUID: identity.UID, Name: name, Email: identity.Email, Role: "user", Active: true})
		if errors.As(err, &e) && e.Status == 409 {
			user, err = s.Users.Get(ctx, id)
		}
		if err != nil {
			return models.User{}, err
		}
	}
	if !user.Active {
		return models.User{}, http.Fail(403, "User is inactive")
	}
	return user, nil
}

// List returns a paginated collection.
func (s *Service) List(ctx context.Context, q url.Values) (models.Page[models.User], error) {
	return s.Users.List(ctx, q)
}

// User retrieves a user by ID.
func (s *Service) User(ctx context.Context, id string) (models.User, error) {
	return s.Users.Get(ctx, id)
}
