package website

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"stocks/models"
	"stocks/utils"
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
		WebsitePaths  Paths
	}{
		AccountTicker: models.LETFAccountTicker(etf),
		Holdings:      etfHoldings,
		Overlaps:      overlaps,
		AccountsMap:   s.metadata.AccountMap,
		WebsitePaths:  s.config.WebsitePaths,
	}
	c.HTML(http.StatusOK, ETFSummaryTemplate, data)
}

func (s *server) fetchETF(c *gin.Context) string {
	etf := c.Param("etf")
	etf = strings.ReplaceAll(etf, ".html", "")
	return etf
}
