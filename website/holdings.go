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

	m := map[models.StockTicker][]models.LETFHolding{}
	var etfHoldings []models.LETFHolding
	etfString := "Custom"
	for index, value := range marketValuesMapFloats {
		if stockName, ok := stocksMap[index]; ok {
			stockTicker := utils.FetchStockTicker(stockName)
			etfHolding := models.LETFHolding{
				TradeDate:         "N/A",
				LETFAccountTicker: models.LETFAccountTicker(etfString),
				LETFDescription:   "Custom Holdings created by customer",
				StockTicker:       stockTicker,
				StockDescription:  s.metadata.StocksMap[stockTicker].StockDescription,
				Shares:            0,
				Price:             0,
				NotionalValue:     0,
				MarketValue:       int64(value),
				PercentContained:  utils.RoundedPercentage(value / totalMarketValue * 100),
				Provider:          "Customer",
			}
			m[stockTicker] = []models.LETFHolding{etfHolding}
			etfHoldings = append(etfHoldings, etfHolding)
		} else {
			log.Panicf("did not find index: %s in the stocksMap; %s", index, debugMessage)
		}
	}

	overlapAnalyses := s.findOverlaps(models.LETFAccountTicker(etfString), m)
	s._renderETF(c, etfString, etfHoldings, overlapAnalyses)
}

func (s *server) findOverlaps(etfName models.LETFAccountTicker, customHoldings map[models.StockTicker][]models.LETFHolding) map[string][]models.LETFOverlapAnalysis {
	var eligibleETFs = map[models.LETFAccountTicker]bool{}
	for stockTicker := range customHoldings {
		stockHoldings, err := s.dependencies.Logger.FetchStock(string(stockTicker))
		utils.PanicErr(err)
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
