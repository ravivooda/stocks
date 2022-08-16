package insights

import (
	"context"
	"stocks/models"
)

type FileName string

type Logger interface {
	LogOverlapAnalysis(analysis models.LETFOverlapAnalysis) (FileName, error)
	LogHoldings(context context.Context, etfName models.LETFAccountTicker, holdings []models.LETFHolding, leverageMappedOverlaps map[string][]models.LETFOverlapAnalysis) (FileName, error)
	FetchHoldings(etfName string) ([]models.LETFHolding, error)
}

type Config struct {
	OverlapsDirectory    string
	ETFHoldingsDirectory string
}
