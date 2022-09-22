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
	etfHoldings, _, err := s.dependencies.Logger.FetchHoldings(etf)
	utils.PanicErr(err)

	overlaps, err := s.dependencies.Logger.FetchOverlaps(etf)
	utils.PanicErr(err)

	s._renderETF(c, etf, etfHoldings, overlaps)
}

func (s *server) _renderETF(c *gin.Context, etf string, etfHoldings []models.LETFHolding, overlaps map[string][]models.LETFOverlapAnalysis) {
	data := struct {
		AccountTicker          models.LETFAccountTicker
		Holdings               []models.LETFHolding
		Overlaps               map[string][]models.LETFOverlapAnalysis
		AccountsMap            map[models.LETFAccountTicker]models.ETFMetadata
		TemplateCustomMetadata TemplateCustomMetadata
	}{
		AccountTicker:          models.LETFAccountTicker(etf),
		Holdings:               etfHoldings,
		Overlaps:               overlaps,
		AccountsMap:            s.metadata.AccountMap,
		TemplateCustomMetadata: s.config.TemplateCustomMetadata,
	}
	c.HTML(http.StatusOK, ETFSummaryTemplate, data)
}

func (s *server) fetchETF(c *gin.Context) string {
	etf := c.Param("etf")
	etf = strings.ReplaceAll(etf, ".html", "")
	return etf
}
