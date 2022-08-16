package website

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"stocks/models"
	"stocks/utils"
	"stocks/website/letf"
	"strings"
)

func (s *server) renderETF(c *gin.Context) {
	etf := s.fetchETF(c)
	etfHoldings, err := s.logger.FetchHoldings(etf)
	utils.PanicErr(err)

	overlaps, err := s.logger.FetchOverlaps(etf)
	utils.PanicErr(err)

	data := struct {
		AccountTicker models.LETFAccountTicker
		Holdings      []models.LETFHolding
		Overlaps      map[string][]models.LETFOverlapAnalysis
		AccountsMap   map[models.LETFAccountTicker][]models.LETFHolding
		WebsitePaths  letf.WebsitePaths
	}{
		AccountTicker: models.LETFAccountTicker(etf),
		Holdings:      etfHoldings,
		Overlaps:      overlaps,
		AccountsMap:   s.accountMap,
		WebsitePaths:  s.config.WebsitePaths,
	}
	c.HTML(http.StatusOK, letf.ETFSummaryTemplate, data)
}

func (s *server) fetchETF(c *gin.Context) string {
	etf := c.Param("etf")
	etf = strings.ReplaceAll(etf, ".html", "")
	return etf
}
