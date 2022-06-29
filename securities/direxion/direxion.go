package direxion

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"stocks/models"
	"stocks/securities"
	"stocks/utils"
	"strconv"
	"strings"
)

type client struct {
}

func (d *client) GetHoldings(_ context.Context, seed models.Seed, etf models.ETF) ([]models.LETFHolding, error) {
	data, err := utils.ReadCSVFromUrl(seed.URL, ',', -1)
	if err != nil {
		return nil, err
	}

	if len(data) <= seed.Header.SkippableLines {
		return nil, errors.New(fmt.Sprintf("got fewer (%d) than expected lines (%d) for seed %+v", len(data), seed.Header.SkippableLines, seed))
	}

	if strings.Join(data[seed.Header.SkippableLines-1], ",") != strings.Join(seed.Header.ExpectedColumns, ",") {
		return nil, errors.New(fmt.Sprintf("columns did not match -> expected: (%s), received: (%s) for seed %+v", seed.Header.ExpectedColumns, data[seed.Header.SkippableLines-1], seed))
	}

	data = data[seed.Header.SkippableLines:]
	data = utils.FilterNonStockRows(data, func(row []string) bool {
		return utils.FetchStockTicker(row[2]) != ""
	})

	var totalSum int64
	for i := 0; i < len(data); i++ {
		totalSum += parseInt(data[i][6])
	}

	var holdings []models.LETFHolding
	for i := 0; i < len(data); i++ {
		holdings = append(holdings, models.LETFHolding{
			TradeDate:         data[i][0],
			LETFAccountTicker: utils.FetchAccountTicker(data[i][1]),
			LETFDescription:   etf.ETFName,
			StockTicker:       utils.FetchStockTicker(data[i][2]),
			StockDescription:  data[i][3],
			Shares:            parseInt(data[i][4]),
			Price:             parseInt(data[i][5]),
			MarketValue:       parseInt(data[i][6]),
			PercentContained:  utils.RoundedPercentage(float64(parseInt(data[i][6])) / float64(totalSum) * 100),
			Provider:          "Direxion",
		})
	}

	return holdings, nil
}

func parseInt(s string) int64 {
	s = strings.Split(s, ".")[0]
	ri, _ := strconv.ParseInt(s, 10, 64)
	return ri
}

func NewClient() (securities.Client, error) {
	return &client{}, nil
}
