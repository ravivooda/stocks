package letf

import "stocks/models"

func (g *generator) logStockSummaryPageToHTML(stockTemplateLoc string, outputFilePath string, ticker string, holdings []models.LETFHolding) (bool, error) {
	data := struct {
		Ticker       string
		Holdings     []models.LETFHolding
		WebsitePaths WebsitePaths
	}{
		Ticker:       ticker,
		Holdings:     holdings,
		WebsitePaths: websitePaths,
	}
	return g.logHTMLWithData(stockTemplateLoc, outputFilePath, data)
}
