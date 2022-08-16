package insights

import (
	"context"
	"stocks/models"
	"stocks/utils"
)

type FileName string

type Logger interface {
	LogOverlapAnalysis(leverage string, analysis models.LETFOverlapAnalysis) (FileName, error)
	LogHoldings(context context.Context, etfName models.LETFAccountTicker, holdings []models.LETFHolding) (FileName, error)
	FetchHoldings(etfName string) ([]models.LETFHolding, error)
	FetchOverlaps(etfName string) (map[string][]models.LETFOverlapAnalysis, error)
}

type Config struct {
	OverlapsDirectory    string
	ETFHoldingsDirectory string
}

type logger struct {
	c Config
}

func NewInsightsLogger(config Config) Logger {
	_, err := utils.MakeDirs([]string{
		config.ETFHoldingsDirectory,
		config.OverlapsDirectory,
	})
	utils.PanicErr(err)
	return &logger{c: config}
}
