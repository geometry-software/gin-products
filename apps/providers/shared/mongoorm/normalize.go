package mongoorm

import (
	"github.com/geometry-software/gin-products/apps/providers/shared/http"
	"github.com/geometry-software/gin-products/apps/providers/shared/models"
	"sort"
)

// Normalize merges and validates duplicate product quantities.
func Normalize(items []models.Quantity) ([]models.Quantity, error) {
	counts := map[string]int64{}
	if len(items) == 0 || len(items) > 100 {
		return nil, http.Fail(400, "Expected 1..100 items")
	}
	for _, v := range items {
		if v.Quantity < 1 || v.Quantity > 1000000 || v.ProductID == "" {
			return nil, http.Fail(400, "Invalid item")
		}
		counts[v.ProductID] += v.Quantity
		if counts[v.ProductID] > 1000000 {
			return nil, http.Fail(400, "Quantity too large")
		}
	}
	ids := make([]string, 0, len(counts))
	for id := range counts {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	result := make([]models.Quantity, 0, len(ids))
	for _, id := range ids {
		result = append(result, models.Quantity{ProductID: id, Quantity: counts[id]})
	}
	return result, nil
}
