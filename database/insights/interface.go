package insights

import (
	"context"
	"stocks/insights/overlap"
	"stocks/models"
	"stocks/utils"
)

type FileName string

type Logger interface {
	LogOverlapAnalysisForHolder(lhs models.LETFAccountTicker, wrappers []OverlapWrapper) (FileName, error)
	LogHoldings(context context.Context, etfName models.LETFAccountTicker, holdings []models.LETFHolding, leverage string) (FileName, error)
	FetchHoldings(etfName string) (etfHoldings []models.LETFHolding, leverage string, err error)
	FetchOverlapsWithoutDetailedOverlaps(etfName string) (map[string][]models.LETFOverlapAnalysis, error)
	FetchOverlapDetails(lhs string, rhs []string) (models.LETFOverlapAnalysis, error)
	LogStocks(ctx context.Context, holdingsWithStockTickerMap map[models.StockTicker][]models.LETFHolding) ([]FileName, error)
	FetchStock(stock string) ([]models.LETFHolding, error)
	HasStock(stock string) (bool, error)
}

type Config struct {
	OverlapsDirectory    string
	ETFHoldingsDirectory string
	StocksDirectory      string
	RootDirectory        string
}

type logger struct {
	c Config
	g overlap.Generator
}

func NewInsightsLogger(config Config, generator overlap.Generator) Logger {
	_, err := utils.MakeDirs([]string{
		config.ETFHoldingsDirectory,
		config.OverlapsDirectory,
		config.StocksDirectory,
	})
	utils.PanicErr(err)
	return &logger{
		c: config,
		g: generator,
	}
}
