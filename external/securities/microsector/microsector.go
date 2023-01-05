package microsector

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"math"
	"os"
	"stocks/external/securities"
	"stocks/models"
	"stocks/utils"
	"strconv"
)

type client struct {
}

func (c client) GetHoldings(_ context.Context, seed models.Seed, etf models.ETF) ([]models.LETFHolding, error) {
	file, err := os.Open(fmt.Sprintf("external/securities/microsector/holdings/%s_Holdings.csv", seed.Ticker))
	defer func(file *os.File) {
		_ = file.Close()
	}(file)
	if err != nil {
		return nil, err
	}
	csvReader := csv.NewReader(file)
	allRows, err := csvReader.ReadAll()
	if err != nil {
		return nil, err
	}

	var rets []models.LETFHolding
	totalPercentage := float64(0)
	for _, row := range allRows {
		parsedPercentage, _ := strconv.ParseFloat(row[2], 64)
		rets = append(rets, models.LETFHolding{
			TradeDate:         utils.TodayDate(),
			LETFAccountTicker: utils.FetchAccountTicker(seed.Ticker),
			LETFDescription:   etf.ETFName,
			StockTicker:       utils.FetchStockTicker(row[1]),
			StockDescription:  row[0],
			Shares:            0,
			Price:             0,
			MarketValue:       0,
			PercentContained:  parsedPercentage,
			Provider:          "MicroSector",
		})
		totalPercentage += parsedPercentage
	}
	if math.Abs(totalPercentage-100) >= 0.1 {
		return nil, errors.New(fmt.Sprintf("total percentage (%f) did not add up to 100 percent for seed %+v", totalPercentage, seed))
	}
	return rets, nil
}

func NewClient() (securities.Client, error) {
	return &client{}, nil
}
