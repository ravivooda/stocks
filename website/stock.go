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

	sort.Slice(holdings, func(i, j int) bool {
		return holdings[i].PercentContained > holdings[j].PercentContained
	})

	data := struct {
		Ticker       string
		Holdings     []models.LETFHolding
		WebsitePaths letf.WebsitePaths
	}{
		Ticker:       stock,
		Holdings:     holdings,
		WebsitePaths: s.config.WebsitePaths,
	}
	c.HTML(http.StatusOK, letf.StockSummaryTemplate, data)
}

func (s *server) fetchStock(c *gin.Context) string {
	etf := c.Param(stockParamKey)
	etf = strings.ReplaceAll(etf, ".html", "")
	return etf
}
