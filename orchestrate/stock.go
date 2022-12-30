package orchestrate

import (
	"context"
	"fmt"
	"stocks/database/insights"
	"stocks/models"
	"stocks/utils"
	"strings"
)

func logStocks(
	ctx context.Context,
	request Request,
	holdingsWithStockTickerMap map[models.StockTicker][]models.LETFHolding,
	tickerMap map[models.LETFAccountTicker][]models.LETFHolding,
	etfMaps map[models.LETFAccountTicker]models.ETF,
) {
	var stockWrappersMap = map[models.StockTicker]insights.StockWrapper{}
	for ticker, holdings := range holdingsWithStockTickerMap {
		var maxCombinations []models.StockCombination
		for i := 1; i < 6; i++ {
			var max = i
			var currentMaxCombination = models.StockCombination{}
			for _, etfHolding := range holdings {
				// TODO: Forcing to create combinations of stocks only for 1x etfs
				if etfMaps[etfHolding.LETFAccountTicker].Leveraged != "1x" {
					continue
				}
				etfHoldings := tickerMap[etfHolding.LETFAccountTicker]
				combination := combinationForTicker(etfHoldings, max, ticker, etfHolding)

				if combination.SummedPercent > currentMaxCombination.SummedPercent {
					currentMaxCombination = combination
				}
			}
			maxCombinations = append(maxCombinations, currentMaxCombination)
		}
		stockWrappersMap[ticker] = insights.StockWrapper{
			Holdings:     holdings,
			Combinations: maxCombinations,
		}
	}
	fmt.Printf("Generating %d stock summaries\n", len(holdingsWithStockTickerMap))
	fileNames, err := request.InsightsLogger.LogStocks(ctx, stockWrappersMap)
	fmt.Printf("Generated files: %s\n", fileNames)
	utils.PanicErr(err)
}

func combinationForTicker(etfHoldings []models.LETFHolding, max int, ticker models.StockTicker, etfHolding models.LETFHolding) models.StockCombination {
	var summed = etfHolding.PercentContained
	var stocks = []models.StockTicker{ticker}
	for _, holding := range etfHoldings {
		if len(stocks) == max {
			break
		}
		if holding.StockTicker == ticker || holding.StockTicker == "" || strings.ToLower(string(holding.StockTicker)) == "cash" {
			continue
		}
		stocks = append(stocks, holding.StockTicker)
		summed += holding.PercentContained
	}
	var combination = models.StockCombination{
		Tickers: stocks,
		ETF: models.ETFMetadata{
			Ticker: etfHolding.LETFAccountTicker,
		},
		SummedPercent: utils.RoundedPercentage(summed),
	}
	return combination
}
