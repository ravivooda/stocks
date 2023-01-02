package direxion

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"stocks/models"
	"stocks/securities"
	"stocks/utils"
	"strings"
)

type client struct {
	config Config
}

type Config struct {
	TemporaryDir string
}

func (d *client) GetHoldings(_ context.Context, seed models.Seed, etf models.ETF) ([]models.LETFHolding, error) {
	data, err := utils.ReadCSVFromUrlWithLocalMasks(seed.URL, d.config.TemporaryDir, ',', -1)
	utils.PanicErr(err)

	if len(data) <= seed.Header.SkippableLines {
		utils.PanicErr(errors.New(fmt.Sprintf("got fewer (%d) than expected lines (%d) for seed %+v", len(data), seed.Header.SkippableLines, seed)))
	}

	if strings.Join(data[seed.Header.SkippableLines-1], ",") != strings.Join(seed.Header.ExpectedColumns, ",") {
		utils.PanicErr(errors.New(fmt.Sprintf("columns did not match -> expected: (%s), received: (%s) for seed %+v", seed.Header.ExpectedColumns, data[seed.Header.SkippableLines-1], seed)))
	}

	data = data[seed.Header.SkippableLines:]
	//data = utils.FilterNonStockRows(data, func(row []string) bool {
	//	return utils.FetchStockTicker(row[2]) != ""
	//})

	var totalPercent = 0.0

	var holdings []models.LETFHolding
	for i := 0; i < len(data); i++ {
		// Skip if it's a SWAP
		d := data[i]
		if strings.ToLower(d[7]) == "swap" {
			continue
		}
		percent := utils.ParseFloat(d[8])
		if percent >= 100 || percent <= -100 {
			panic(fmt.Sprintf("Percent was beyong comprehensible: %f, %+v, %+v, %+v\n", percent, data, holdings, seed))
		}
		totalPercent += percent
		holdings = append(holdings, models.LETFHolding{
			TradeDate:         d[0],
			LETFAccountTicker: utils.FetchAccountTicker(d[1]),
			LETFDescription:   etf.ETFName,
			StockTicker:       utils.FetchStockTicker(d[2]),
			StockDescription:  d[3],
			Shares:            utils.ParseInt(d[4]),
			Price:             utils.ParseInt(d[5]),
			MarketValue:       utils.ParseInt(d[6]),
			PercentContained:  percent,
			Provider:          "Direxion",
		})
	}

	if totalPercent < 70 || totalPercent > 125 {
		panic(fmt.Sprintf("holdings did not add up to 100 for %+v, %+v, %f\n", holdings, seed, totalPercent))
	}

	return holdings, nil
}

func NewClient(config Config) (securities.Client, error) {
	_, err := utils.MakeDirs([]string{config.TemporaryDir})
	utils.PanicErr(err)
	return &client{
		config: config,
	}, nil
}
