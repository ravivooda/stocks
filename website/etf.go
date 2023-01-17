package website

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"stocks/external/stocks/alphavantage"
	"stocks/models"
	"stocks/utils"
	"strings"
)

func (s *server) renderETF(c *gin.Context) {
	etf := s.fetchETF(c)
	etfHoldings, _, err := s.dependencies.Logger.FetchHoldings(etf)
	utils.PanicErr(err)

	overlaps, err := s.dependencies.Logger.FetchOverlapsWithoutDetailedOverlaps(etf)
	utils.PanicErr(err)

	s._renderETF(c, etf, etfHoldings, overlaps)
}

func (s *server) _renderETF(c *gin.Context, etf string, etfHoldings []models.LETFHolding, overlaps map[string][]models.LETFOverlapAnalysis) {
	totalOverlapAnalyses := 0
	for _, analyses := range overlaps {
		totalOverlapAnalyses += len(analyses)
	}

	latestDate, latestData, linear5DaysData, err := s.fetchStockTradingData(etf, 5) // TODO: Hardcoded 5 days split data

	data := struct {
		AccountTicker            models.LETFAccountTicker
		Holdings                 []models.LETFHolding
		Overlaps                 map[string][]models.LETFOverlapAnalysis
		AccountsMap              map[models.LETFAccountTicker]models.ETFMetadata
		OverlapsTotalCount       int
		TemplateCustomMetadata   TemplateCustomMetadata
		TotalProvidersCount      int
		ShouldRenderAlphaVantage bool
		LatestData               alphavantage.DailyData
		LatestDate               string
		Top10Percentage          float64
		ChartData                ChartData
	}{
		AccountTicker:            models.LETFAccountTicker(etf),
		Holdings:                 etfHoldings,
		Overlaps:                 overlaps,
		AccountsMap:              s.metadata.AccountMap,
		OverlapsTotalCount:       totalOverlapAnalyses,
		TemplateCustomMetadata:   s.metadata.TemplateCustomMetadata,
		TotalProvidersCount:      len(s.metadata.ProvidersMap),
		ShouldRenderAlphaVantage: err == nil,
		LatestData:               latestData,
		LatestDate:               latestDate,
		Top10Percentage:          s.top10HoldingsPercentage(10, etfHoldings),
		ChartData: ChartData{
			Ticker:                 etf,
			LinearDailyData:        linear5DaysData,
			TaxLossCalculationData: s.generateTaxLossCalculationData(etf, linear5DaysData, overlaps),
		},
	}
	c.HTML(http.StatusOK, ETFSummaryTemplate, data)
}

func (s *server) top10HoldingsPercentage(max int, etfHoldings []models.LETFHolding) float64 {
	top10HoldingsPercentage := 0.0
	for i := 0; i < max; i++ {
		top10HoldingsPercentage += etfHoldings[i].PercentContained
	}
	return utils.RoundedPercentage(top10HoldingsPercentage)
}

func (s *server) fetchETF(c *gin.Context) string {
	etf := c.Param("etf")
	etf = strings.ReplaceAll(etf, ".html", "")
	return etf
}
