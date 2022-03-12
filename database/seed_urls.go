package database

import (
	"context"
	"imports/models"
)

type dumbDatabase struct {
}

func (i dumbDatabase) ListSeeds(context context.Context) ([]models.Seed, error) {
	expectedColumns := []string{
		"TradeDate",
		"AccountTicker",
		"StockTicker",
		"SecurityDescription",
		"Shares",
		"Price",
		"MarketValue",
	}
	header := models.Header{
		SkippableLines:  4,
		ExpectedColumns: expectedColumns,
		OutstandingShares: models.OutstandingShares{
			LineNumber: 2,
			Prefix:     "Shares Outstanding:",
		},
	}
	return []models.Seed{
		{
			URL:    "https://www.direxion.com/holdings/TECL.csv",
			Ticker: "TECL",
			Header: header,
		},
		{
			URL:    "https://www.direxion.com/holdings/TECS.csv",
			Ticker: "TECS",
			Header: header,
		},
	}, nil
}

func NewDumbDatabase() DB {
	return &dumbDatabase{}
}
