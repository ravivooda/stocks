package website

import (
	"math"
	"stocks/external/stocks/alphavantage"
	"stocks/utils"
)

func (s *server) generateTaxLossCalculationData(data []alphavantage.LinearTimeSeriesDaily) TaxLossCalculationData {
	start, end := data[0], data[len(data)-1]
	startPrice, endPrice := utils.ParseFloat(start.DailyPrice), utils.ParseFloat(end.DailyPrice)
	beginPortfolioValue := 10000.00
	endPortfolioValue := (beginPortfolioValue / startPrice) * endPrice
	return TaxLossCalculationData{
		Begin: alphavantage.LinearTimeSeriesDaily{
			Date:       start.Date,
			DailyPrice: renderLargeNumbers(int(beginPortfolioValue)),
		},
		Today: alphavantage.LinearTimeSeriesDaily{
			Date:       end.Date,
			DailyPrice: renderLargeNumbers(int(endPortfolioValue)),
		},
		IsHarvesteable: startPrice >= endPrice,
		ChangePrice:    renderLargeNumbers(int(math.Abs(beginPortfolioValue - endPortfolioValue))),
		Swappable:      "ABC",
	}
}
