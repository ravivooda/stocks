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
	return []models.Seed{
		//{
		//	URL:             "https://www.direxion.com/holdings/TECL.csv",
		//	Ticker:          "TECL",
		//	SkippableLines:  4,
		//	ExpectedColumns: expectedColumns,
		//},
		{
			URL:             "https://www.direxion.com/holdings/TECS.csv",
			Ticker:          "TECS",
			SkippableLines:  4,
			ExpectedColumns: expectedColumns,
		},
	}, nil
}

func NewDumbDatabase() DB {
	return &dumbDatabase{}
}
