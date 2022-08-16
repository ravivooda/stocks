package website

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"sort"
	"stocks/models"
	"stocks/utils"
	"stocks/website/letf"
	"strings"
)

func (s *server) renderStock(c *gin.Context) {
	stock := s.fetchStock(c)
	holdings, err := s.logger.FetchStock(stock)
	utils.PanicErr(err)

	mappedHoldings := s.mappedHoldings(holdings)

	for leverage, letfHoldings := range mappedHoldings {
		sort.Slice(letfHoldings, func(i, j int) bool {
			return letfHoldings[i].PercentContained > letfHoldings[j].PercentContained
		})
		mappedHoldings[leverage] = letfHoldings
	}

	data := struct {
		Ticker         string
		MappedHoldings map[string][]models.LETFHolding
		WebsitePaths   letf.WebsitePaths
	}{
		Ticker:         stock,
		MappedHoldings: mappedHoldings,
		WebsitePaths:   s.config.WebsitePaths,
	}
	c.HTML(http.StatusOK, letf.StockSummaryTemplate, data)
}

func (s *server) mappedHoldings(holdings []models.LETFHolding) map[string][]models.LETFHolding {
	mappedHoldings := map[string][]models.LETFHolding{}
	for _, holding := range holdings {
		leverage := s.etfsMaps[holding.LETFAccountTicker].Leveraged
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
