package letf

import (
	"context"
	"fmt"
	"stocks/models"
	"stocks/utils"
)

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

func (g *generator) GenerateStock(_ context.Context, stockTicker models.StockTicker, letfHoldings []models.LETFHolding) {
	escapedTickerString := string(stockTicker)
	stockSummaryFilePath := fmt.Sprintf("%s/%s.html", g.stockSummariesFileRoot, escapedTickerString)
	_, err := g.logStockSummaryPageToHTML(stockSummaryTemplateLoc, stockSummaryFilePath, escapedTickerString, letfHoldings)
	utils.PanicErr(err)
}
