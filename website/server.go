package website

import (
	"context"
	"stocks/database/insights"
	"stocks/models"
	"stocks/website/letf"
)

type Server interface {
	StartServing(ctx context.Context) error
}

type Config struct {
	InsightsConfig insights.Config
	WebsitePaths   letf.WebsitePaths
}

type Metadata struct {
	AccountMap map[models.LETFAccountTicker][]models.LETFHolding
	EtfsMap    map[models.LETFAccountTicker]models.ETF
	StocksMap  map[models.StockTicker]models.StockMetadata
}

type server struct {
	config   Config
	logger   insights.Logger
	metadata Metadata
}

func New(
	config Config,
	logger insights.Logger,
	metadata Metadata,
) Server {
	return &server{
		config:   config,
		logger:   logger,
		metadata: metadata,
	}
}
