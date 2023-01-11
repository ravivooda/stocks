package website

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"sort"
	"stocks/external/stocks/alphavantage"
	"stocks/models"
	"stocks/utils"
	"strings"
)

func (s *server) renderStock(c *gin.Context) {
	stock := s.fetchStock(c)
	stockWrapper, err := s.dependencies.Logger.FetchStock(stock)
	utils.PanicErr(err)

	mappedHoldings := s.mappedHoldings(stockWrapper.Holdings)

	for leverage, letfHoldings := range mappedHoldings {
		sort.Slice(letfHoldings, func(i, j int) bool {
			return letfHoldings[i].PercentContained > letfHoldings[j].PercentContained
		})
		mappedHoldings[leverage] = letfHoldings
	}

	latestDate, latestData, linear5DaysData, err := s.fetchStockTradingData(stock, 5) // TODO: Hardcoded 10 days split data

	data := struct {
		Ticker                   string
		Description              string
		MappedHoldings           map[string][]models.LETFHolding
		Combinations             []models.StockCombination
		TemplateCustomMetadata   TemplateCustomMetadata
		TotalETFsCount           int
		TotalProvidersCount      int
		ShouldRenderAlphaVantage bool
		LatestData               alphavantage.DailyData
		LatestDate               string
		LinearDailyData          []alphavantage.LinearTimeSeriesDaily
	}{
		Ticker:                   stock,
		Description:              stockWrapper.Holdings[0].StockDescription, //TODO: Hardcoded stock description
		MappedHoldings:           mappedHoldings,
		TemplateCustomMetadata:   s.metadata.TemplateCustomMetadata,
		Combinations:             stockWrapper.Combinations,
		TotalETFsCount:           len(stockWrapper.Holdings),
		TotalProvidersCount:      len(s.metadata.ProvidersMap),
		ShouldRenderAlphaVantage: err == nil,
		LatestData:               latestData,
		LatestDate:               latestDate,
		LinearDailyData:          linear5DaysData,
	}
	c.HTML(http.StatusOK, StockSummaryTemplate, data)
}

func (s *server) fetchStockTradingData(ticker string, X int) (
	string,
	alphavantage.DailyData,
	[]alphavantage.LinearTimeSeriesDaily,
	error,
) {
	// See if we can fetch data from alpha vantage about the stock
	tradingData, err := s.dependencies.AlphaVantage.FetchStockTradingData(ticker)
	if err != nil {
		fmt.Printf("Error when fetching trading data for %s, error is: %s\n", ticker, err)
		return "", alphavantage.DailyData{}, nil, err
	}

	latestDate := tradingData.LatestDate()
	if k, ok := tradingData.TimeSeriesDaily[latestDate]; ok {
		return latestDate, k, tradingData.SplitByXDates(X), nil
	}
	return "", alphavantage.DailyData{}, nil, errors.New(fmt.Sprintf("Hmm, did not find lastData for %s when combing through %+v\n", latestDate, tradingData))
}

func (s *server) mappedHoldings(holdings []models.LETFHolding) map[string][]models.LETFHolding {
	mappedHoldings := map[string][]models.LETFHolding{}
	for _, holding := range holdings {
		leverage := s.metadata.AccountMap[holding.LETFAccountTicker].Leveraged
		a := mappedHoldings[leverage]
		if a == nil {
			a = []models.LETFHolding{}
		}
		a = append(a, holding)
		mappedHoldings[leverage] = a
	}
	return mappedHoldings
}

func (s *server) fetchStock(c *gin.Context) string {
	etf := c.Param(stockParamKey)
	etf = strings.ReplaceAll(etf, ".html", "")
	return etf
}
