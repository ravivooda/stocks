package direxion

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"imports/models"
	"io"
	"net/http"
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

	if len(data) <= seed.SkippableLines {
		return nil, errors.New(fmt.Sprintf("got fewer (%d) than expected lines (%d)", len(data), seed.SkippableLines))
	}

	if strings.Join(data[seed.SkippableLines-1], ",") != strings.Join(seed.ExpectedColumns, ",") {
		return nil, errors.New(fmt.Sprintf("columns did not match -> expected: (%s), received: (%s)", seed.ExpectedColumns, data[seed.SkippableLines-1]))
	}

	var totalSum float64
	for i := seed.SkippableLines; i < len(data); i++ {
		totalSum += parseFloat(data[i][6])
	}

	var holdings []models.Holding
	for i := seed.SkippableLines; i < len(data); i++ {
		mv := parseFloat(data[i][6])
		holdings = append(holdings, models.Holding{
			TradeDate:     data[i][0],
			AccountTicker: data[i][1],
			StockTicker:   data[i][2],
			Description:   data[i][3],
			Shares:        parseInt(data[i][4]),
			Price:         parseFloat(data[i][5]),
			MarketValue:   mv,
			Percent:       mv / totalSum * 100,
		})

		totalSum += mv
	}

	return holdings, nil
}

func parseInt(s string) int64 {
	ri, _ := strconv.ParseInt(s, 10, 64)
	return ri
}

func parseFloat(s string) float64 {
	rii, _ := strconv.ParseFloat(s, 32)
	return rii
}

func NewDirexionClient() (Client, error) {
	return &direxionClient{}, nil
}
