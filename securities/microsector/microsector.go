package microsector

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"stocks/models"
	"stocks/securities"
	"time"
)

type client struct {
}

func (c client) GetHoldings(_ context.Context, seed models.Seed) ([]models.LETFHolding, error) {
	file, err := os.Open(fmt.Sprintf("securities/microsector/holdings/%s_Holdings.csv", seed.Ticker))
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
	for _, row := range allRows {
		rets = append(rets, models.LETFHolding{
			TradeDate:         time.Now().Format("01-02-2006"),
			LETFAccountTicker: models.LETFAccountTicker(seed.Ticker),
			StockTicker:       models.StockTicker(row[1]),
			Description:       row[0],
			Shares:            0,
			Price:             0,
			MarketValue:       0,
			Percent:           10,
		})
	}
	return rets, nil
}

func NewClient() (securities.Client, error) {
	return &client{}, nil
}
