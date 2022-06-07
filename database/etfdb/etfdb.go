package etfdb

import (
	"context"
	"github.com/pkg/errors"
	"stocks/models"
	"stocks/utils"
	"strings"
)

type client struct {
	config Config
}

type Generator interface {
	ListETFs(ctx context.Context) ([]models.ETF, error)
}

const (
	csvPath         = "database/etfdb/etfs_details_type_fund_flow.csv"
	expectedHeaders = "Symbol,ETF Name,Asset Class,Total Assets ,YTD Price Change,Avg. Daily Volume,Previous Closing Price,1-Day Change,Inverse,Leveraged,1 Week,1 Month,1 Year,3 Year,5 Year,YTD FF,1 Week FF,4 Week FF,1 Year FF,3 Year FF,5 Year FF,ETF Database Category,Inception,ER,Commission Free,Annual Dividend Rate,Dividend Date,Dividend,Annual Dividend Yield %,P/E Ratio,Beta,# of Holdings,% In Top 10,ST Cap Gain Rate,LT Cap Gain Rate,Tax Form,Lower Bollinger,Upper Bollinger,Support 1,Resistance 1,RSI,Liquidity Rating,Expenses Rating,Returns Rating,Volatility Rating,Dividend Rating,Concentration Rating,ESG Score,ESG Score Peer Percentile (%),ESG Score Global Percentile (%),Carbon Intensity (Tons of CO2e / $M Sales),SRI Exclusion Criteria (%),Sustainable Impact Solutions (%)"
)

func (c *client) ListETFs(_ context.Context) ([]models.ETF, error) {
	rows, err := utils.ReadCSVFromLocalFile(csvPath)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, errors.Errorf("found empty rows when reading from %s", csvPath)
	}

	if x := strings.Join(rows[0], ","); x != expectedHeaders {
		return nil, errors.Errorf("headers did not match expected. \n\tObserved Headers: %s,\n\tExpected Headers: %s", x, expectedHeaders)
	}
	rows = rows[1:]
	var rets []models.ETF
	for _, row := range rows {
		rets = append(rets, models.GenerateETFFromStrings(row))
	}
	return rets, nil
}

type Config struct {
}

func New(config Config) Generator {
	return &client{config: config}
}
