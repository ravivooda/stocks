package securities

import (
	"context"
	"stocks/models"
)

// Client API
type Client interface {
	GetHoldings(ctx context.Context, seed models.Seed) ([]models.LETFHolding, error)
}
