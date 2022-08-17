package website

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"stocks/models"
)

func (s *server) renderIndex(c *gin.Context) {
	var data = struct {
		TotalProvider     int
		TotalSeeds        int
		TotalStockTickers int
		Providers         map[models.Provider]models.ProviderMetadata
		Stocks            map[string][]models.StockTicker
		WebsitePaths      Paths
	}{
		TotalProvider:     len(s.metadata.ProvidersMap),
		TotalSeeds:        len(s.metadata.AccountMap),
		TotalStockTickers: len(s.metadata.StocksMap),
		Providers:         s.metadata.ProvidersMap,
		Stocks:            s.welcomeStocksRenderingMap(),
		WebsitePaths:      s.config.WebsitePaths,
	}

	c.HTML(http.StatusOK, WelcomeTemplate, data)
}

func (s *server) welcomeStocksRenderingMap() map[string][]models.StockTicker {
	var groupedStocks = map[string][]models.StockTicker{}
	for ticker := range s.metadata.StocksMap {
		s := "unknown"
		if len(ticker) > 0 {
			s = string(ticker[0:1])
		}
		a := groupedStocks[s]
		if a == nil {
			a = []models.StockTicker{}
		}
		a = append(a, ticker)
		groupedStocks[s] = a
	}
	return groupedStocks
}
