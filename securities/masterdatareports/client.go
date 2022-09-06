package masterdatareports

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"math"
	"stocks/models"
	"stocks/securities"
	"stocks/utils"
	"strconv"
	"strings"
	"time"
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
		for _, holding := range holdings {
			if math.IsNaN(holding.PercentContained) {
				utils.PanicErr(errors.New(fmt.Sprintf("Unexpected NaN for %v with holdings: %v", etf, holdings)))
			}
		}
		return holdings, nil
	}
	return nil, errors.Errorf("unable to find mapping for holding: %s", etf.Symbol)
}

const (
	expectedHeaders = "Sponsor,Composite Ticker,Composite Name,Constituent Ticker,Constituent Name,Weighting,Identifier,Date,Location,Exchange,Total Shares Held,Notional Value,Market Value,Sponsor Sector,Last Trade,Currency,BloombergSymbol,BloombergExchange,NAICSSector,NAICSSubIndustry,Coupon,Maturity,Rating,Type,SharesOutstanding,MarketCap,Earnings,PE Ratio,Face,FIGI,TimeZone,DividendAmt,XDate,DividendYield,RIC,IssueType,NAICSSector,NAICSIndustry,NAICSSubIndustry,CUSIP,ISIN,FIGI"
)

var (
	skippables = map[string]int{
		"r1sm2":     1,
		"parent":    1,
		"weight":    1,
		"ticker":    1,
		"rank":      1,
		"weighting": 1,
	}
	// SkippedOwners should contain lowercase string
	skippedOwners = map[string]int{
		"overlay":   1,
		"qraft":     1,
		"oshares":   1,
		"proshares": 1,
		"hartford":  1,
		"equbot":    1,
	}
	skippingTickers = map[string]int{
		"EMAG": 1,
		"PFUT": 1,
		"PLDR": 1,
		"RESD": 1,
		"STLG": 1,
	}
)

func New(config Config) (Client, error) {
	records := loadData(config)

	if len(records) == 0 {
		return nil, errors.Errorf("found empty rows when trying to load csv from url: %s", config.HoldingsCSVURL)
	}

	if x := strings.Join(records[0], ","); x != expectedHeaders {
		return nil, errors.Errorf("headers did not match expected. \n\tObserved Headers: %s,\n\tExpected Headers: %s", x, expectedHeaders)
	}

	records = records[1:]
	var parsedHoldings = map[models.LETFAccountTicker][]models.LETFHolding{}
	var differentOwners = map[string]map[string]bool{}
	for _, record := range records {
		m := differentOwners[record[0]]
		if m == nil {
			m = map[string]bool{}
		}
		m[record[1]] = true
		differentOwners[record[0]] = m
		if _, ok := skippables[strings.ToLower(record[5])]; ok {
			continue
		} else if _, ok := skippables[strings.ToLower(record[4])]; ok {
			continue
		} else if _, ok := skippedOwners[strings.ToLower(record[0])]; ok {
			continue
		} else if _, ok := skippingTickers[strings.ToUpper(record[1])]; ok {
			continue
		}
		var holding = parse(record)
		holdings := parsedHoldings[holding.LETFAccountTicker]
		if holdings == nil {
			holdings = []models.LETFHolding{}
		}
		holdings = append(holdings, holding)
		parsedHoldings[holding.LETFAccountTicker] = holdings
		if math.IsNaN(holding.PercentContained) {
			utils.PanicErr(errors.New("Unexpected NaN"))
		}
	}

	for s, m := range differentOwners {
		fmt.Printf("Provider: %s (%d), %+v\n", s, len(m), m)
	}

	var mappedHoldings = map[models.LETFAccountTicker][]models.LETFHolding{}
	debuggingTicker := models.LETFAccountTicker("WBIF")
	ShouldBeSkipped := map[models.LETFAccountTicker]bool{}
	for ticker, holdings := range parsedHoldings {
		totalMarketValue := int64(0)
		totalledPercentage := float64(0)
		if ticker == debuggingTicker {
			fmt.Println("Asdasdasdasd")
		}
		for _, holding := range holdings {
			totalMarketValue += holding.MarketValue
			if holding.PercentContained < 0 && ticker == debuggingTicker {
				fmt.Println("asdasdasdsd")
			}
			totalledPercentage += holding.PercentContained
		}

		if totalledPercentage > 0.5 && totalledPercentage <= 1.1 {
			var holdingsWithPercentage []models.LETFHolding
			for _, holding := range holdings {
				holding.PercentContained = utils.RoundedPercentage(holding.PercentContained * 100)
				holdingsWithPercentage = append(holdingsWithPercentage, holding)
			}
			mappedHoldings[ticker] = holdingsWithPercentage
			continue
		}

		var holdingsWithPercentage []models.LETFHolding
		for _, holding := range holdings {
			if totalledPercentage != 0 {
				holding.PercentContained = utils.RoundedPercentage(holding.PercentContained)
			} else {
				if totalMarketValue == 0 {
					//panic(errors.New(fmt.Sprintf("total market value is 0 for : %v", holdings)))
					ShouldBeSkipped[ticker] = true
				}
				holding.PercentContained = utils.RoundedPercentage(float64(holding.MarketValue) / float64(totalMarketValue) * 100.00)
			}
			holdingsWithPercentage = append(holdingsWithPercentage, holding)
		}
		mappedHoldings[ticker] = holdingsWithPercentage
	}

	fmt.Printf("going to skip: %v\n", ShouldBeSkipped)

	return &client{
		config:   config,
		holdings: mappedHoldings,
	}, nil
}

func loadData(config Config) [][]string {
	local := false
	if local {
		return loadLocalCacheFile()
	} else {
		return loadRemoteData(config)
	}
}

func loadLocalCacheFile() (records [][]string) {
	const hardcodedCSVLocation = "securities/masterdatareports/Backup/ETFData42.csv"
	records, err := utils.RetryFetching(func() ([][]string, error) {
		return utils.ReadCSVFromLocalFile(hardcodedCSVLocation)
	}, 3, 0)
	utils.PanicErr(err)
	fmt.Printf("From local file, number of records: %d\n", len(records))
	return records
}

func loadRemoteData(config Config) [][]string {
	defer utils.Elapsed("Master Data Reports Loading")()
	fmt.Printf("Fetching holdings from %s\n", config.HoldingsCSVURL)
	records, err := utils.RetryFetching(func() ([][]string, error) {
		return utils.ReadCSVFromUrl(config.HoldingsCSVURL, ',', -1)
	}, 3, 10*time.Second)
	fmt.Printf("From remote file, number of records: %d\n", len(records))
	utils.PanicErr(err)
	return records
}

func parse(record []string) models.LETFHolding {
	percentContained, err := strconv.ParseFloat(record[5], 64)
	if err != nil {
		percentContained = 0
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
	ticker := ""
	if strings.TrimSpace(record[16]) != "" && strings.Count(record[16], ":") == 1 {
		ticker = strings.Split(record[16], ":")[0]
	} else {
		ticker = record[3]
	}
	if math.IsNaN(percentContained) {
		utils.PanicErr(errors.New("Unexpected NaN"))
	}
	return models.LETFHolding{
		TradeDate:         record[7],
		LETFAccountTicker: utils.FetchAccountTicker(record[1]),
		LETFDescription:   record[2],
		StockTicker:       utils.FetchStockTicker(ticker),
		StockDescription:  record[4],
		//TODO: Fill the following information from the csv
		Shares:           int64(shares),
		Price:            0,
		MarketValue:      int64(marketValue),
		NotionalValue:    notionalValue,
		PercentContained: percentContained * 100.00, // The CSV reports weights summing up to 1
		Provider:         fmt.Sprintf("Master Data Reports: %s", record[0]),
	}
}
