package template4

import (
	"fmt"
	"stocks/models"
	"stocks/utils"
)

var Headers = []string{
	"Fund Ticker",
	"Security Identifier",
	"Identifier",
	"Shares",
	"Weight",
	"Name",
	"Contract Expiry Date",
	"Sector",
	"$ Value",
	"Date",
}

func Parse(data [][]string, seed models.Seed, etf models.ETF) ([]models.LETFHolding, error) {
	data = data[seed.Header.SkippableLines:]

	totalPercent := 0.0
	var holdings []models.LETFHolding
	for _, datum := range data {
		if !isValid(datum) {
			continue
		}

		holding := holding(datum, etf)
		totalPercent += holding.PercentContained
		holdings = append(holdings, holding)
	}

	if totalPercent < 95 || totalPercent > 105 {
		panic(fmt.Sprintf("[template 4] holdings did not add up to 100 for %+v, %+v, %f\n", holdings, seed, totalPercent))
		//fmt.Printf("[Template 3] invesco error: holdings did not add up to 100 for %+v, got %f\n", seed, totalPercent)
	}

	return holdings, nil
}

func holding(datum []string, etf models.ETF) models.LETFHolding {
	return models.LETFHolding{
		TradeDate:         datum[9],
		LETFAccountTicker: utils.FetchAccountTicker(datum[0]),
		LETFDescription:   etf.ETFName,
		StockTicker:       utils.FetchStockTicker(datum[2]),
		StockDescription:  datum[5],
		Shares:            utils.ParseInt(datum[3]),
		Price:             0,
		NotionalValue:     0,
		MarketValue:       utils.ParseInt(datum[8]),
		PercentContained:  utils.ParseFloat(datum[4]),
		Provider:          models.Invesco,
	}
}

func isValid(datum []string) bool {
	return true
}
