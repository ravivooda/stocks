package letf

import "stocks/models"

func (g *generator) logWelcomePageToHTML(welcomePageTemplateLoc, outputFilePath string, request Request) (bool, error) {
	var mapped = map[string]map[models.LETFAccountTicker]bool{}
	for ticker, holdings := range request.Letfs {
		providerMap := mapped[holdings[0].Provider]
		if providerMap == nil {
			providerMap = map[models.LETFAccountTicker]bool{}
		}
		providerMap[ticker] = true
		mapped[holdings[0].Provider] = providerMap
	}
	var groupedStocks = map[string][]models.StockTicker{}
	for ticker := range request.StocksMap {
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
	var data = struct {
		TotalProvider int
		TotalSeeds    int
		Providers     map[string]map[models.LETFAccountTicker]bool
		Stocks        map[string][]models.StockTicker
		WebsitePaths  WebsitePaths
	}{
		TotalProvider: len(mapped),
		TotalSeeds:    len(request.Letfs),
		Providers:     mapped,
		Stocks:        groupedStocks,
		WebsitePaths:  websitePaths,
	}
	return g.logHTMLWithData(welcomePageTemplateLoc, outputFilePath, data)
}
