package website

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"stocks/models"
	"stocks/utils"
	"strings"
)

const sep = ","

func (s *server) renderOverlap(c *gin.Context) {
	holder, holdees := s.fetchOverlapParam(c)
	analysis, err := s.dependencies.Logger.FetchOverlapDetails(holder, holdees)
	utils.PanicErr(err)

	data := struct {
		Analysis               models.LETFOverlapAnalysis
		StocksMap              map[models.StockTicker]models.StockMetadata
		ETFsMap                map[models.LETFAccountTicker]models.ETFMetadata
		TemplateCustomMetadata TemplateCustomMetadata
	}{
		Analysis:               analysis,
		StocksMap:              s.metadata.StocksMap,
		ETFsMap:                s.metadata.AccountMap,
		TemplateCustomMetadata: s.metadata.TemplateCustomMetadata,
	}

	c.HTML(http.StatusOK, OverlapTemplate, data)
}

func (s *server) fetchOverlapParam(c *gin.Context) (string, []string) {
	// TODO: Handle error gracefully for query lookup in overlaps
	lhs, _ := c.GetQuery(overlapKeyLHS)
	rhs, _ := c.GetQuery(overlapKeyRHS)
	return lhs, strings.Split(rhs, sep)
}
