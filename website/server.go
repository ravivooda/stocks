package website

import (
	"context"
	"stocks/database/insights"
	"stocks/external/stocks/alphavantage"
	"stocks/insights/overlap"
	"stocks/models"
	"time"
)

type Server interface {
	StartServing(ctx context.Context, killIn time.Duration) error
}

type Dependencies struct {
	Logger       insights.Logger
	Generator    overlap.Generator
	AlphaVantage alphavantage.Client
}

type Config struct {
	InsightsConfig insights.Config
}

type Metadata struct {
	AccountMap             map[models.LETFAccountTicker]models.ETFMetadata
	StocksMap              map[models.StockTicker]models.StockMetadata
	ProvidersMap           map[models.Provider]models.ProviderMetadata
	TemplateCustomMetadata TemplateCustomMetadata
}

type AutoCompleteMetadata struct {
	StocksMap  []models.StockTicker
	AccountMap []models.LETFAccountTicker
}

type server struct {
	config       Config
	dependencies Dependencies
	metadata     Metadata
}

func New(
	config Config,
	dependencies Dependencies,
	metadata Metadata,
) Server {
	return &server{
		config:       config,
		dependencies: dependencies,
		metadata:     metadata,
	}
}

type Paths struct {
	LETFSummary      string
	StockSummary     string
	Overlaps         string
	TemplatesRootDir string
}
