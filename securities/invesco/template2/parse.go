package template2

import (
	"fmt"
	"stocks/models"
	"stocks/utils"
)

var Headers = []string{
	"Fund Ticker",
	"Holding Ticker",
	"Security Identifier",
	"Name",
	"CouponRate",
	"MaturityDate",
	"Effective Date",
	"Next_Call_Date",
	"rating",
	"Shares/Par Value",
	"MarketValue",
	"PercentageOfFund",
	"PositionDate",
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
		panic(fmt.Sprintf("[template 2] holdings did not add up to 100 for %+v, %+v, %f\n", holdings, seed, totalPercent))
		//fmt.Printf("invesco error: holdings did not add up to 100 for %+v, got %f\n", seed, totalPercent)
	}

	return holdings, nil
}

func holding(datum []string, etf models.ETF) models.LETFHolding {
	return models.LETFHolding{
		TradeDate:         datum[12],
		LETFAccountTicker: utils.FetchAccountTicker(datum[0]),
		LETFDescription:   etf.ETFName,
		StockTicker:       utils.FetchStockTicker(datum[1]),
		StockDescription:  datum[3],
		Shares:            utils.ParseInt(datum[9]),
		Price:             0,
		NotionalValue:     0,
		MarketValue:       utils.ParseInt(datum[10]),
		PercentContained:  utils.ParseFloat(datum[11]),
		Provider:          models.Invesco,
	}
}

func isValid(datum []string) bool {
	return true
}
