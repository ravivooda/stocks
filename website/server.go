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

type server struct {
	config     Config
	logger     insights.Logger
	accountMap map[models.LETFAccountTicker][]models.LETFHolding
}

func New(config Config, logger insights.Logger, accountMap map[models.LETFAccountTicker][]models.LETFHolding) Server {
	return &server{
		config:     config,
		logger:     logger,
		accountMap: accountMap,
	}
}
