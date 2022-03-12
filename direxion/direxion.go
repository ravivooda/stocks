package direxion

import (
	"context"
	"encoding/csv"
	"imports/models"
	"io"
	"net/http"
	"strconv"
)

// DirexionClient API
type DirexionClient interface {
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
	reader.Comma = ';'
	data, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var holdings []models.Holding
	for i := seed.SkippableLines; i < len(data); i++ {
		holdings = append(holdings, models.Holding{
			TradeDate:     data[i][0],
			AccountTicker: data[i][1],
			StockTicker:   data[i][2],
			Description:   data[i][3],
			Shares:        parseInt(data[i][4]),
			Price:         parseFloat(data[i][5]),
			MarketValue:   parseFloat(data[i][6]),
			Percent:       float32(parseFloat(data[i][7])),
		})
	}

	return holdings, nil
}

func parseInt(s string) int64 {
	ri, _ := strconv.ParseInt(s, 10, 64)
	return ri
}

func parseFloat(s string) float64 {
	rii, _ := strconv.ParseFloat(s, 64)
	return rii
}

func NewDirexionClient() (DirexionClient, error) {
	return &direxionClient{}, nil
}
