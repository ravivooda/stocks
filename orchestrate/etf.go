package orchestrate

import (
	"context"
	"fmt"
	"stocks/database/insights"
	"stocks/models"
	"stocks/utils"
)

func logHoldings(
	ctx context.Context,
	logger insights.Logger,
	holdingsWithAccountTickerMap map[models.LETFAccountTicker][]models.LETFHolding,
	etfsMap map[models.LETFAccountTicker]models.ETF,
) {
	for ticker, holdings := range holdingsWithAccountTickerMap {
		fileName, err := logger.LogHoldings(ctx, ticker, holdings, etfsMap[ticker].Leveraged)
		utils.PanicErr(err)
		fmt.Printf("wrote the holdings to %s\n", fileName)
	}
}
