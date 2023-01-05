package website

import (
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
	stockTicker := utils.FetchStockTicker(stock)
	stockWrapper, err := s.dependencies.Logger.FetchStock(stock)
	utils.PanicErr(err)

	mappedHoldings := s.mappedHoldings(stockWrapper.Holdings)

	for leverage, letfHoldings := range mappedHoldings {
		sort.Slice(letfHoldings, func(i, j int) bool {
			return letfHoldings[i].PercentContained > letfHoldings[j].PercentContained
		})
		mappedHoldings[leverage] = letfHoldings
	}

	latestDate, latestData := s.fetchStockTradingData(stockTicker)

	data := struct {
		Ticker                 string
		Description            string
		MappedHoldings         map[string][]models.LETFHolding
		Combinations           []models.StockCombination
		TemplateCustomMetadata TemplateCustomMetadata
		TotalETFsCount         int
		TotalProvidersCount    int
		LatestData             alphavantage.DailyData
		LatestDate             string
	}{
		Ticker:                 stock,
		Description:            stockWrapper.Holdings[0].StockDescription, //TODO: Hardcoded stock description
		MappedHoldings:         mappedHoldings,
		TemplateCustomMetadata: s.metadata.TemplateCustomMetadata,
		Combinations:           stockWrapper.Combinations,
		TotalETFsCount:         len(stockWrapper.Holdings),
		TotalProvidersCount:    len(s.metadata.ProvidersMap),
		LatestData:             latestData,
		LatestDate:             latestDate,
	}
	c.HTML(http.StatusOK, StockSummaryTemplate, data)
}

func (s *server) fetchStockTradingData(stockTicker models.StockTicker) (string, alphavantage.DailyData) {
	// See if we can fetch data from alpha vantage about the stock
	tradingData, err := s.dependencies.AlphaVantage.FetchStockTradingData(stockTicker)
	if err != nil {
		fmt.Printf("Error when fetching trading data for %s, error is: %s\n", stockTicker, err)
	} else {
		latestDate := tradingData.LatestDate()
		if k, ok := tradingData.TimeSeriesDaily[latestDate]; ok {
			return latestDate, k
		}
		fmt.Printf("Hmm, did not find lastData for %s when combing through %+v\n", latestDate, tradingData)
	}
	return "", alphavantage.DailyData{}
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
