package proshares

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"reflect"
	"stocks/database"
	"stocks/models"
	"stocks/securities"
	"stocks/utils"
	"strconv"
	"strings"
)

type SeedProvider interface {
	database.DB
	securities.Client
}

type client struct {
	seeds          []models.Seed
	cachedHoldings map[models.LETFAccountTicker][]models.LETFHolding
}

func (c *client) ListSeeds(_ context.Context) ([]models.Seed, error) {
	return c.seeds, nil
}

func (c *client) GetHoldings(_ context.Context, seed models.Seed) ([]models.LETFHolding, error) {
	if retVal, ok := c.cachedHoldings[models.LETFAccountTicker(seed.Ticker)]; ok {
		return retVal, nil
	}
	return nil, errors.New(fmt.Sprintf("did not find holdings for %s in %v", seed.Ticker, keysFromMap(c.cachedHoldings)))
}

func keysFromMap(holdings map[models.LETFAccountTicker][]models.LETFHolding) []models.LETFAccountTicker {
	var rets []models.LETFAccountTicker
	for ticker := range holdings {
		rets = append(rets, ticker)
	}
	return rets
}

type Config struct {
	CSVURL          string
	SkipLines       int      `mapstructure:"skip_lines"`
	ExpectedColumns []string `mapstructure:"expected_columns"`
}

func New(config Config) (SeedProvider, error) {
	csvFromUrl, err := utils.ReadCSVFromUrl(config.CSVURL, ',', -1)
	if err != nil {
		return nil, errors.Wrapf(err, "Found error when parsing csv from url %v", config.CSVURL)
	}

	if !reflect.DeepEqual(utils.Trimmed(csvFromUrl[config.SkipLines-1]), config.ExpectedColumns) {
		return nil, errors.New(fmt.Sprintf("Expected Columns: %v did not exactly match observed: %v", config.ExpectedColumns, csvFromUrl[config.SkipLines-1]))
	}

	mappedHoldings := createMapWithAccountTicker(config, csvFromUrl)

	seeds, cachedHoldings, err := parseMappedHoldings(mappedHoldings)
	if err != nil {
		return nil, err
	}

	return &client{cachedHoldings: cachedHoldings, seeds: seeds}, nil
}

func createMapWithAccountTicker(config Config, csvFromUrl [][]string) map[models.LETFAccountTicker][][]string {
	mappedHoldings := map[models.LETFAccountTicker][][]string{}
	for i := config.SkipLines; i < len(csvFromUrl); i++ {
		rowTicker := utils.FetchAccountTicker(csvFromUrl[i][0])
		var groupedArray = mappedHoldings[rowTicker]
		if groupedArray == nil {
			groupedArray = [][]string{}
		}
		groupedArray = append(groupedArray, csvFromUrl[i])
		mappedHoldings[rowTicker] = groupedArray
	}
	return mappedHoldings
}

func parseMappedHoldings(mappedHoldings map[models.LETFAccountTicker][][]string) ([]models.Seed, map[models.LETFAccountTicker][]models.LETFHolding, error) {
	var cachedHoldings = map[models.LETFAccountTicker][]models.LETFHolding{}
	var seeds []models.Seed
	for rowTicker, groupedArray := range mappedHoldings {
		if _, found := ignoreHoldings[rowTicker]; found {
			// TODO: Holdings ignored because of problems in the csv from ProShares
			continue
		}
		if _, found := knowinglyIgnoredIssues[rowTicker]; found {
			// TODO: Holdings ignored because of problems in the csv from ProShares
			continue
		}
		seeds = append(seeds, models.Seed{
			URL:      "",
			Ticker:   string(rowTicker),
			Header:   models.Header{},
			Provider: models.ProShares,
		})
		groupedArray = utils.FilterNonStockRows(groupedArray, func(row []string) bool {
			return getStockTicker(row) != ""
		})
		totalMarketValue := int64(0)
		for _, csvHoldingRow := range groupedArray {
			marketValue, err := getMarketValue(csvHoldingRow, ".")
			if err != nil {
				return nil, nil, errors.Wrapf(err, "parsing market value from the values: %+v", csvHoldingRow)
			}
			totalMarketValue += marketValue
		}
		var holdingsArray = cachedHoldings[rowTicker]
		if holdingsArray == nil {
			holdingsArray = []models.LETFHolding{}
		}

		for _, csvHoldingRow := range groupedArray {
			stockTicker := getStockTicker(csvHoldingRow)
			if stockTicker == "" {
				continue
			}
			shares, err := strconv.ParseInt(splitForIntString(csvHoldingRow[7], "."), 10, 0)
			if err != nil {
				return nil, nil, err
			}
			marketValue, err := getMarketValue(csvHoldingRow, ".")
			if err != nil {
				return nil, nil, err
			}
			holdingsArray = append(holdingsArray, models.LETFHolding{
				TradeDate:         utils.TodayDate(),
				LETFAccountTicker: rowTicker,
				StockTicker:       stockTicker,
				LETFDescription:   csvHoldingRow[1],
				StockDescription:  csvHoldingRow[4],
				Shares:            shares,
				Price:             0,
				MarketValue:       marketValue,
				PercentContained:  utils.RoundedPercentage(float64(marketValue) / float64(totalMarketValue) * 100),
				Provider:          "ProShares",
			})
		}
		cachedHoldings[rowTicker] = holdingsArray
	}
	return seeds, cachedHoldings, nil
}

func getStockTicker(csvHoldingRow []string) models.StockTicker {
	return utils.FetchStockTicker(csvHoldingRow[2])
}

func getMarketValue(row []string, separator string) (int64, error) {
	if strings.TrimSpace(row[9]) != "" {
		return strconv.ParseInt(splitForIntString(row[9], separator), 10, 0)
	}
	return strconv.ParseInt(splitForIntString(row[8], separator), 10, 0)
}

func splitForIntString(value string, separator string) string {
	return strings.Split(strings.TrimSpace(value), separator)[0]
}
