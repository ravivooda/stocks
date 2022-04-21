package direxion

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"stocks/models"
	"stocks/securities"
	"strconv"
	"strings"
)

type client struct {
}

func (d *client) GetHoldings(_ context.Context, seed models.Seed) ([]models.LETFHolding, error) {
	resp, err := http.Get(seed.URL)
	if err != nil {
		return nil, fmt.Errorf("fetching %+v returned err: %w", seed, err)
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			print(err)
		}
	}(resp.Body)

	reader := csv.NewReader(resp.Body)
	reader.Comma = ','
	reader.FieldsPerRecord = -1
	data, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("reading data for %+v returned err: %w", seed, err)
	}

	if len(data) <= seed.Header.SkippableLines {
		return nil, errors.New(fmt.Sprintf("got fewer (%d) than expected lines (%d) for seed %+v", len(data), seed.Header.SkippableLines, seed))
	}

	if strings.Join(data[seed.Header.SkippableLines-1], ",") != strings.Join(seed.Header.ExpectedColumns, ",") {
		return nil, errors.New(fmt.Sprintf("columns did not match -> expected: (%s), received: (%s) for seed %+v", seed.Header.ExpectedColumns, data[seed.Header.SkippableLines-1], seed))
	}

	var totalSum int64
	for i := seed.Header.SkippableLines; i < len(data); i++ {
		if shouldIgnoreCashFund(data[i][3]) {
			continue
		}
		totalSum += parseInt(data[i][6])
	}

	//fmt.Println(totalSum)
	var totalPercent float64
	var holdings []models.LETFHolding
	for i := seed.Header.SkippableLines; i < len(data); i++ {
		if shouldIgnoreCashFund(data[i][3]) {
			continue
		}
		holdings = append(holdings, models.LETFHolding{
			TradeDate:         data[i][0],
			LETFAccountTicker: models.LETFAccountTicker(data[i][1]),
			StockTicker:       models.StockTicker(data[i][2]),
			Description:       data[i][3],
			Shares:            parseInt(data[i][4]),
			Price:             parseInt(data[i][5]),
			MarketValue:       parseInt(data[i][6]),
			Percent:           float64(parseInt(data[i][6])) / float64(totalSum) * 100,
		})
		totalPercent += float64(parseInt(data[i][6])) / float64(totalSum) * 100
	}

	if math.Abs(totalPercent-100) >= 0.1 {
		return nil, errors.New(fmt.Sprintf("total percentage (%f) did not add up to 100 percent for seed %+v", totalPercent, seed))
	}

	return holdings, nil
}

func shouldIgnoreCashFund(name string) bool {
	if name == "TECHNOLOGY SELECT SECTOR INDEX SWAP" || name == "ICE SEMICONDUCTOR INDEX SWAP" || name == "DREYFUS GOVT CASH MGMT" {
		return true
	}
	return false
}

func parseInt(s string) int64 {
	s = strings.Split(s, ".")[0]
	ri, _ := strconv.ParseInt(s, 10, 64)
	return ri
}

func NewClient() (securities.Client, error) {
	return &client{}, nil
}
