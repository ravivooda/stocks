package securities

import (
	"context"
	"stocks/database"
	"stocks/models"
)

// Client API
type Client interface {
	GetHoldings(ctx context.Context, seed models.Seed, etf models.ETF) ([]models.LETFHolding, error)
}

type ClientV2 interface {
	GetHoldings(ctx context.Context, etf models.ETF) ([]models.LETFHolding, error)
}

type SeedProvider interface {
	database.DB
	Client
}
