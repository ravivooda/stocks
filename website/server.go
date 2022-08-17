package website

import (
	"context"
	"stocks/database/insights"
	"stocks/models"
)

type Server interface {
	StartServing(ctx context.Context) error
}

type Config struct {
	InsightsConfig insights.Config
	WebsitePaths   Paths
}

type Metadata struct {
	AccountMap   map[models.LETFAccountTicker][]models.LETFHolding
	EtfsMap      map[models.LETFAccountTicker]models.ETF
	StocksMap    map[models.StockTicker]models.StockMetadata
	ProvidersMap map[models.Provider]models.ProviderMetadata
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

type Paths struct {
	LETFSummary      string
	StockSummary     string
	Overlaps         string
	TemplatesRootDir string
}
