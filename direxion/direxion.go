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
	"strconv"
	"strings"
)

// Client API
type Client interface {
	GetHoldings(ctx context.Context, seed models.Seed) ([]models.Holding, error)
}

type direxionClient struct {
}

func (d *direxionClient) GetHoldings(ctx context.Context, seed models.Seed) ([]models.Holding, error) {
	resp, err := http.Get(seed.URL)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	if len(data) <= seed.Header.SkippableLines {
		return nil, errors.New(fmt.Sprintf("got fewer (%d) than expected lines (%d)", len(data), seed.Header.SkippableLines))
	}

	if strings.Join(data[seed.Header.SkippableLines-1], ",") != strings.Join(seed.Header.ExpectedColumns, ",") {
		return nil, errors.New(fmt.Sprintf("columns did not match -> expected: (%s), received: (%s)", seed.Header.ExpectedColumns, data[seed.Header.SkippableLines-1]))
	}

	var totalSum int64
	for i := seed.Header.SkippableLines; i < len(data); i++ {
		totalSum += parseInt(data[i][6])
	}

	//fmt.Println(totalSum)
	var totalPercent float64
	var holdings []models.Holding
	for i := seed.Header.SkippableLines; i < len(data); i++ {
		holdings = append(holdings, models.Holding{
			TradeDate:     data[i][0],
			AccountTicker: data[i][1],
			StockTicker:   data[i][2],
			Description:   data[i][3],
			Shares:        parseInt(data[i][4]),
			Price:         parseInt(data[i][5]),
			MarketValue:   parseInt(data[i][6]),
			Percent:       float64(parseInt(data[i][6])) / float64(totalSum) * 100,
		})
		totalPercent += float64(parseInt(data[i][6])) / float64(totalSum) * 100
	}

	if math.Abs(totalPercent-100) >= 0.1 {
		return nil, errors.New(fmt.Sprintf("total percentage (%f) did not add up to 100%", totalPercent))
	}

	return holdings, nil
}

func parseInt(s string) int64 {
	s = strings.Split(s, ".")[0]
	ri, _ := strconv.ParseInt(s, 10, 64)
	return ri
}

func NewDirexionClient() (Client, error) {
	return &direxionClient{}, nil
}
