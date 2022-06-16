package masterdatareports

import (
	"context"
	"github.com/pkg/errors"
	"stocks/models"
	"stocks/securities"
	"stocks/utils"
	"strconv"
	"strings"
)

type Client interface {
	securities.ClientV2
	Count() int
}

type Config struct {
	HoldingsCSVURL string `mapstructure:"holdings_csv_url"`
}

type client struct {
	config   Config
	holdings map[models.LETFAccountTicker][]models.LETFHolding
}

func (c *client) Count() int {
	return len(c.holdings)
}

func (c *client) GetHoldings(_ context.Context, etf models.ETF) ([]models.LETFHolding, error) {
	if holdings, ok := c.holdings[etf.Symbol]; ok {
		return holdings, nil
	}
	return nil, errors.Errorf("unable to find mapping for holding: %s", etf.Symbol)
}

const (
	expectedHeaders      = "Sponsor,Composite Ticker,Composite Name,Constituent Ticker,Constituent Name,Weighting,Identifier,Date,Location,Exchange,Total Shares Held,Notional Value,Market Value,Sponsor Sector,Last Trade,Currency,BloombergSymbol,BloombergExchange,NAICSSector,NAICSSubIndustry,Coupon,Maturity,Rating,Type,SharesOutstanding,MarketCap,Earnings,PE Ratio,Face,FIGI,TimeZone,DividendAmt,XDate,DividendYield,RIC,IssueType,NAICSSector,NAICSIndustry,NAICSSubIndustry,CUSIP,ISIN,FIGI"
	hardcodedCSVLocation = "securities/masterdatareports/ETFData42.csv"
)

func New(config Config) (Client, error) {
	records, err := utils.ReadCSVFromLocalFile(hardcodedCSVLocation)
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return nil, errors.Errorf("found empty rows when trying to load csv from url: %s", config.HoldingsCSVURL)
	}

	if x := strings.Join(records[0], ","); x != expectedHeaders {
		return nil, errors.Errorf("headers did not match expected. \n\tObserved Headers: %s,\n\tExpected Headers: %s", x, expectedHeaders)
	}

	records = records[1:]
	var parsedHoldings = map[models.LETFAccountTicker][]models.LETFHolding{}
	for _, record := range records {
		if strings.ToLower(record[5]) == "parent" ||
			strings.ToLower(record[5]) == "weight" {
			continue
		}
		var holding = parse(record)
		holdings := parsedHoldings[holding.LETFAccountTicker]
		if holdings == nil {
			holdings = []models.LETFHolding{}
		}
		holdings = append(holdings, holding)
		parsedHoldings[holding.LETFAccountTicker] = holdings
	}

	var mappedHoldings = map[models.LETFAccountTicker][]models.LETFHolding{}
	for ticker, holdings := range parsedHoldings {
		totalMarketValue := int64(0)
		totalledPercentage := float64(0)
		for _, holding := range holdings {
			totalMarketValue += holding.MarketValue
			totalledPercentage += holding.PercentContained
		}

		if totalledPercentage > 0.5 && totalledPercentage <= 1.0 {
			var holdingsWithPercentage []models.LETFHolding
			for _, holding := range holdings {
				holding.PercentContained *= 100
				holdingsWithPercentage = append(holdingsWithPercentage, holding)
			}
			mappedHoldings[ticker] = holdingsWithPercentage
			continue
		} else if totalledPercentage > 70 && totalledPercentage <= 101 {
			mappedHoldings[ticker] = holdings
			continue
		}

		var holdingsWithPercentage []models.LETFHolding
		for _, holding := range holdings {
			holding.PercentContained = utils.RoundedPercentage(float64(holding.MarketValue) / float64(totalMarketValue) * 100.00)
			holdingsWithPercentage = append(holdingsWithPercentage, holding)
		}
		mappedHoldings[ticker] = holdingsWithPercentage
	}

	return &client{
		config:   config,
		holdings: mappedHoldings,
	}, nil
}

func parse(record []string) models.LETFHolding {
	percentContained, err := strconv.ParseFloat(record[5], 64)
	if err != nil {
		percentContained = -1
	}
	marketValue, err := strconv.ParseFloat(record[12], 64)
	if err != nil {
		marketValue = 0
	}
	notionalValue, err := strconv.ParseFloat(record[11], 64)
	if err != nil {
		notionalValue = 0
	}
	shares, err := strconv.ParseFloat(record[10], 64)
	if err != nil {
		shares = 0
	}
	return models.LETFHolding{
		TradeDate:         record[7],
		LETFAccountTicker: utils.FetchAccountTicker(record[1]),
		LETFDescription:   record[2],
		StockTicker:       utils.FetchStockTicker(record[3]),
		StockDescription:  record[4],
		//TODO: Fill the following information from the csv
		Shares:           int64(shares),
		Price:            0,
		MarketValue:      int64(marketValue),
		NotionalValue:    notionalValue,
		PercentContained: percentContained * 100.00, // The CSV reports weights summing up to 1
		Provider:         models.MasterDataReports,
	}
}
