package providers

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// Kind identifies an infrastructure provider hosted by this service.
type Kind string

const (
	// Stock is the transactional product stock provider.
	Stock Kind = "stock"
)

// All returns the registered provider kinds in startup order.
func All() []Kind { return []Kind{Stock} }

// Register mounts each provider's routes on the service router.
func Register(r *gin.Engine, db *mongo.Database) error {
	for _, kind := range All() {
		switch kind {
		case Stock:
			registerStock(r, db)
		default:
			return fmt.Errorf("unregistered provider %q", kind)
		}
	}
	return nil
}
