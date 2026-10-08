// Package mongoorm provides typed document mapping and an HTTP repository port.
// MongoDB connections and BSON mapping live exclusively in the adapters service.
package mongoorm

import (
	"context"
	stdhttp "net/http"
	"net/url"

	"github.com/geometry-software/gin-products/apps/providers/shared/http"
	"github.com/geometry-software/gin-products/apps/providers/shared/models"
)

// Repository defines typed persistence operations.
type Repository[T any] interface {
	Create(context.Context, T) (T, error)
	Get(context.Context, string) (T, error)
	List(context.Context, url.Values) (models.Page[T], error)
	Replace(context.Context, string, T) (T, error)
	Delete(context.Context, string, int64) error
}

// RemoteRepository calls the adapters service for typed persistence.
type RemoteRepository[T any] struct{ BaseURL, Caller, Collection string }

func (r RemoteRepository[T]) path(id string) string {
	p := r.BaseURL + "/internal/mongoorm/" + url.PathEscape(r.Collection)
	if id != "" {
		p += "/" + url.PathEscape(id)
	}
	return p
}

// Create validates and persists a new record.
func (r RemoteRepository[T]) Create(ctx context.Context, v T) (out T, err error) {
	err = http.Call(ctx, stdhttp.MethodPost, r.path(""), r.Caller, v, &out)
	return
}

// Get returns a record by ID.
func (r RemoteRepository[T]) Get(ctx context.Context, id string) (out T, err error) {
	err = http.Call(ctx, stdhttp.MethodGet, r.path(id), r.Caller, nil, &out)
	return
}

// List returns a paginated collection.
func (r RemoteRepository[T]) List(ctx context.Context, q url.Values) (out models.Page[T], err error) {
	err = http.Call(ctx, stdhttp.MethodGet, r.path("")+"?"+q.Encode(), r.Caller, nil, &out)
	return
}

// Replace replaces a record using optimistic version checking.
func (r RemoteRepository[T]) Replace(ctx context.Context, id string, v T) (out T, err error) {
	err = http.Call(ctx, stdhttp.MethodPut, r.path(id), r.Caller, v, &out)
	return
}

// Delete removes a record by its ID.
func (r RemoteRepository[T]) Delete(ctx context.Context, id string, version int64) error {
	return http.Call(ctx, stdhttp.MethodDelete, r.path(id), r.Caller, map[string]int64{"version": version}, nil)
}
