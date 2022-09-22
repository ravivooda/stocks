package website

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"stocks/models"
)

func (s *server) renderAllETFs(c *gin.Context) {
	data := s.allRenderingData()
	c.HTML(http.StatusOK, listAllETFsTemplate, data)
}

func (s *server) renderAllStocks(c *gin.Context) {
	data := s.allRenderingData()

	c.HTML(http.StatusOK, listAllStocksTemplate, data)
}

type allData struct {
	TotalProvider          int
	TotalSeeds             int
	TotalStockTickers      int
	Providers              map[models.Provider]models.ProviderMetadata
	Stocks                 map[string][]models.StockTicker
	TemplateCustomMetadata TemplateCustomMetadata
}

func (s *server) allRenderingData() allData {
	var data = allData{
		TotalProvider:          len(s.metadata.ProvidersMap),
		TotalSeeds:             len(s.metadata.AccountMap),
		TotalStockTickers:      len(s.metadata.StocksMap),
		Providers:              s.metadata.ProvidersMap,
		Stocks:                 s.stocksRenderingMap(),
		TemplateCustomMetadata: s.config.TemplateCustomMetadata,
	}
	return data
}

func (s *server) stocksRenderingMap() map[string][]models.StockTicker {
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
