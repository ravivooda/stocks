package website

import (
	"math"
	"stocks/external/stocks/alphavantage"
	"stocks/models"
	"stocks/utils"
)

func (s *server) generateTaxLossCalculationData(
	data []alphavantage.LinearTimeSeriesDaily,
	swappables []models.LETFAccountTicker,
) TaxLossCalculationData {
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
		IsHarvestable: startPrice >= endPrice,
		ChangePrice:   renderLargeNumbers(int(math.Abs(beginPortfolioValue - endPortfolioValue))),
		Swappables:    swappables,
	}
}
