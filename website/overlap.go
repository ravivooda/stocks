package website

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"stocks/models"
	"stocks/utils"
	"strings"
)

const sep = "_"

func (s *server) renderOverlap(c *gin.Context) {
	holder, holdees := s.fetchOverlapParam(c)
	analysis, err := s.dependencies.Logger.FetchOverlap(holder, holdees)
	utils.PanicErr(err)

	data := struct {
		Analysis     models.LETFOverlapAnalysis
		StocksMap    map[models.StockTicker]models.StockMetadata
		WebsitePaths Paths
	}{
		Analysis:     analysis,
		StocksMap:    s.metadata.StocksMap,
		WebsitePaths: s.config.WebsitePaths,
	}

	c.HTML(http.StatusOK, OverlapTemplate, data)
}

func (s *server) fetchOverlapParam(c *gin.Context) (string, string) {
	param := c.Param(overlapParam)
	param = strings.ReplaceAll(param, ".html", "")
	holds := strings.SplitN(param, sep, 2)
	if len(holds) != 2 {
		panic(fmt.Sprintf("expected splittable by _ to 2 segments but did not find anyin %s", param))
	}
	return holds[0], holds[1]
}
