package website

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"stocks/models"
	"stocks/utils"
	"strconv"
)

func (s *server) renderFindOverlapsInputHTML(c *gin.Context) {
	c.HTML(http.StatusOK, findOverlapsInputTemplate, struct {
		TemplateCustomMetadata TemplateCustomMetadata
		TotalProvidersCount    int
		TotalStocksCount       int
		TotalETFsCount         int
		TotalOverlaps          int
	}{
		TemplateCustomMetadata: s.metadata.TemplateCustomMetadata,
		TotalProvidersCount:    len(s.metadata.ProvidersMap),
		TotalStocksCount:       len(s.metadata.StocksMap),
		TotalETFsCount:         len(s.metadata.AccountMap),
		TotalOverlaps:          123456,
	})
}

func (s *server) findOverlapsForCustomHoldings(c *gin.Context) {
	stocksMap := c.PostFormMap("stocks")
	marketValuesMapStrings := c.PostFormMap("market_values")
	marketValuesMapFloats := map[string]float64{}
	totalMarketValue := float64(0)
	debugMessage := fmt.Sprintf("StocksMap = %v, MarketValuesMapStrings = %v, MarketValuesMapFloats = %v", stocksMap, marketValuesMapStrings, marketValuesMapFloats)
	for index, value := range marketValuesMapStrings {
		float, err := strconv.ParseFloat(value, 64)
		utils.PanicErrWithExtraMessage(err, fmt.Sprintf(";%s", debugMessage))
		marketValuesMapFloats[index] = float
		totalMarketValue += float
	}

	var etfHoldings []models.LETFHolding
	etfString := "Custom"
	didNotFindMatches := map[string]float64{}
	for index, value := range marketValuesMapFloats {
		if stockName, ok := stocksMap[index]; ok {
			stockExists, _ := s.dependencies.Logger.HasStock(stockName)
			if stockExists {
				_, etfHolding := s.createStockHolding(stockName, etfString, value, totalMarketValue)
				etfHoldings = append(etfHoldings, etfHolding)
			} else {
				didNotFindMatches[stockName] = value
			}
		} else {
			log.Panicf("did not find index: %s in the stocksMap; %s", index, debugMessage)
		}
	}

	// For all the not found, see if we can find an ETF match instead
	for stockName, value := range didNotFindMatches {
		holdings, _, err := s.dependencies.Logger.FetchHoldings(stockName)
		if err != nil {
			continue
		}
		for i, holding := range holdings {
			holdings[i].StockDescription = fmt.Sprintf("%s (in %s)", holdings[i].StockDescription, stockName)
			holdings[i].PercentContained = utils.RoundedPercentage(value * holding.PercentContained / totalMarketValue)
		}
		etfHoldings, _ = utils.MergeHoldings(etfHoldings, holdings)
		delete(didNotFindMatches, stockName)
	}

	// For all the remaining, just add them as unsupported
	for stockName, value := range didNotFindMatches {
		_, letfHolding := s.createStockHolding(stockName, etfString, value, totalMarketValue)
		etfHoldings = append(etfHoldings, letfHolding)
	}

	m := utils.MapLETFHoldingsWithStockTicker(etfHoldings)

	overlapAnalyses := s.findOverlaps(models.LETFAccountTicker(etfString), m)
	s._renderETF(c, etfString, etfHoldings, overlapAnalyses)
}

func (s *server) createStockHolding(stockName string, etfString string, value float64, totalMarketValue float64) (models.StockTicker, models.LETFHolding) {
	stockTicker := utils.FetchStockTicker(stockName)
	etfHolding := models.LETFHolding{
		TradeDate:         "N/A",
		LETFAccountTicker: models.LETFAccountTicker(etfString),
		LETFDescription:   "Custom Holdings created by customer",
		StockTicker:       stockTicker,
		StockDescription:  s.metadata.StocksMap[stockTicker].StockDescription,
		MarketValue:       int64(value),
		PercentContained:  utils.RoundedPercentage(value / totalMarketValue * 100),
		Provider:          "Customer",
	}
	return stockTicker, etfHolding
}

func (s *server) findOverlaps(etfName models.LETFAccountTicker, customHoldings map[models.StockTicker][]models.LETFHolding) map[string][]models.LETFOverlapAnalysis {
	var eligibleETFs = map[models.LETFAccountTicker]bool{}
	for stockTicker := range customHoldings {
		stockHoldings, err := s.dependencies.Logger.FetchStock(string(stockTicker))
		if err != nil {
			continue
		}
		for _, stockHolding := range stockHoldings {
			eligibleETFs[stockHolding.LETFAccountTicker] = true
		}
	}

	var overlapAnalysis = map[string][]models.LETFOverlapAnalysis{}
	for letfAccountTicker := range eligibleETFs {
		letfHoldings, leverage, err := s.dependencies.Logger.FetchHoldings(string(letfAccountTicker))
		utils.PanicErr(err)
		overlapPercentage, overlaps := s.dependencies.Generator.Compare(customHoldings, utils.MapLETFHoldingsWithStockTicker(letfHoldings))
		overlapsArray := overlapAnalysis[leverage]
		if overlapsArray == nil {
			overlapsArray = []models.LETFOverlapAnalysis{}
		}
		overlapsArray = append(overlapsArray, models.LETFOverlapAnalysis{
			LETFHolder:        etfName,
			LETFHoldees:       []models.LETFAccountTicker{letfAccountTicker},
			OverlapPercentage: overlapPercentage,
			DetailedOverlap:   &overlaps,
		})
		overlapAnalysis[leverage] = overlapsArray
	}
	return utils.SortOverlapsWithinLeverage(overlapAnalysis)
}
