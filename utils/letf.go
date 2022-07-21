package utils

import (
	"fmt"
	"math"
	"math/big"
	"stocks/models"
	"strconv"
	"strings"
	"time"
)

var knownAliases = map[string]string{
	"FB":   "META",
	"BLOC": "SQ",
}

func cleanAndMap(input string) string {
	input = strings.ReplaceAll(input, "/", "_")
	trimmedInput := strings.ToUpper(strings.TrimSpace(input))
	if aliasStockName, ok := knownAliases[trimmedInput]; ok {
		trimmedInput = aliasStockName
	}
	// TODO: Handle / in ticker in a better way than the hack below
	return strings.ReplaceAll(trimmedInput, "/", "_")
}

func FetchStockTicker(input string) models.StockTicker {
	input = strings.ReplaceAll(input, " ", "_")
	input = strings.ReplaceAll(input, ".", "_")
	input = strings.ReplaceAll(input, "-", "_")
	input = strings.ReplaceAll(input, "_US", "")
	input = strings.ReplaceAll(input, "_UQ", "")
	input = strings.ReplaceAll(input, "_UN", "")
	input = strings.ReplaceAll(input, "__", "_")
	input = strings.ReplaceAll(input, "_", " ")
	if len(input) > 1 && input[len(input)-1:] == "_" {
		input = input[:len(input)-2]
	}
	return models.StockTicker(cleanAndMap(input))
}

func FetchAccountTicker(input string) models.LETFAccountTicker {
	return models.LETFAccountTicker(cleanAndMap(input))
}

func MapLETFHoldingsWithAccountTicker(input []models.LETFHolding) map[models.LETFAccountTicker][]models.LETFHolding {
	var rets = map[models.LETFAccountTicker][]models.LETFHolding{}
	for _, holding := range input {
		var s []models.LETFHolding
		if f, ok := rets[holding.LETFAccountTicker]; ok {
			s = f
		}

		s = append(s, holding)
		rets[holding.LETFAccountTicker] = s
	}
	return rets
}

func MapLETFHoldingsWithStockTicker(holdings []models.LETFHolding) map[models.StockTicker][]models.LETFHolding {
	holdingsMap := make(map[models.StockTicker][]models.LETFHolding)
	for _, holding := range holdings {
		var holdingsArray = holdingsMap[holding.StockTicker]
		if holdingsArray == nil {
			holdingsArray = []models.LETFHolding{}
		}
		holdingsArray = append(holdingsArray, holding)
		holdingsMap[holding.StockTicker] = holdingsArray
	}
	return holdingsMap
}

func MapLETFAnalysisWithAccountTicker(analysis []models.LETFOverlapAnalysis) map[models.LETFAccountTicker][]models.LETFOverlapAnalysis {
	analysisMap := map[models.LETFAccountTicker][]models.LETFOverlapAnalysis{}
	for _, overlapAnalysis := range analysis {
		var arr []models.LETFOverlapAnalysis
		if elem, ok := analysisMap[overlapAnalysis.LETFHolder]; ok {
			arr = elem
		}
		arr = append(arr, overlapAnalysis)
		analysisMap[overlapAnalysis.LETFHolder] = arr
	}
	return analysisMap
}

func MappedLETFS(etfs []models.ETF) map[models.LETFAccountTicker]models.ETF {
	var s = map[models.LETFAccountTicker]models.ETF{}
	for _, etf := range etfs {
		s[etf.Symbol] = etf
	}
	return s
}

func GenerateETFFromStrings(input []string) models.ETF {
	return models.ETF{
		Symbol:                        FetchAccountTicker(input[0]),
		ETFName:                       input[1],
		AssetClass:                    input[2],
		TotalAssets:                   parseBigInt(input[3]),
		YTDPriceChange:                parseFloat(input[4]),
		AvgDailyVolume:                parseBigInt(input[5]),
		PreviousClosingPrice:          parseBigInt(input[6]),
		OneDayChange:                  parseFloat(input[7]),
		Inverse:                       parseInversed(input[8]),
		Leveraged:                     parseLeveraged(input[9]),
		OneWeek:                       parseFloat(input[10]),
		OneMonth:                      parseFloat(input[11]),
		OneYear:                       parseFloat(input[12]),
		ThreeYear:                     parseFloat(input[13]),
		FiveYear:                      parseFloat(input[14]),
		YTDFF:                         parseBigInt(input[15]),
		OneWeekFF:                     parseBigInt(input[16]),
		FourWeekFF:                    parseBigInt(input[17]),
		OneYearFF:                     parseBigInt(input[18]),
		ThreeYearFF:                   parseBigInt(input[19]),
		FiveYearFF:                    parseBigInt(input[20]),
		ETFDatabaseCategory:           input[21],
		Inception:                     parseTime(input[22]),
		ER:                            parseFloat(input[23]),
		CommissionFree:                input[24],
		AnnualDividendRate:            parseFloat(input[25]),
		DividendDate:                  parseTime(input[26]),
		Dividend:                      parseFloat(input[27]),
		AnnualDividendYieldPercentage: parseFloat(input[28]),
		PERatio:                       parseFloat(input[29]),
		Beta:                          parseFloat(input[30]),
		NumberOfHoldings:              parseInt(input[31]),
		PercentageInTop10:             parseFloat(input[32]),
		STCapGainRate:                 parseFloat(input[33]),
		LTCapGainRate:                 parseFloat(input[34]),
		TaxForm:                       input[35],
		LowerBollinger:                parseFloat(input[36]),
		UpperBollinger:                parseFloat(input[37]),
		Support1:                      parseFloat(input[38]),
		Resistance1:                   parseFloat(input[39]),
		RSI:                           parseFloat(input[40]),
		LiquidityRating:               input[41],
		ExpensesRating:                input[42],
		ReturnsRating:                 input[43],
		VolatilityRating:              input[44],
		DividendRating:                input[45],
		ConcentrationRating:           input[46],
		ESGScore:                      parseFloat(input[47]),
		ESGScorePeerPercentile:        parseFloat(input[48]),
		ESGScoreGlobalPercentile:      parseFloat(input[49]),
		CarbonIntensity:               parseFloat(input[50]),
		SRIExclusionCriteria:          parseFloat(input[51]),
		SustainableImpactSolutions:    parseFloat(input[52]),
	}
}

func parseInversed(input string) bool {
	return strings.ToLower(input) == "yes"
}

func parseInt(s string) int {
	s = cleanNumberStrings(s)
	s = strings.Split(s, ".")[0]
	i, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		panic(err)
	}
	return int(i)
}

func parseTime(s string) *time.Time {
	if strings.ToLower(s) == "n/a" {
		return nil
	}
	parse, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return &parse
}

func parseLeveraged(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "false" || s == "" {
		return "1x"
	}
	return s
}

func parseFloat(s string) float64 {
	if strings.ToLower(s) == "n/a" {
		return 0
	}
	s = cleanNumberStrings(s)
	float, err := strconv.ParseFloat(s, 64)
	if err != nil {
		panic(err)
	}
	return math.Round(float*100) / 100
}

func parseBigInt(s string) big.Int {
	n := new(big.Int)
	s = cleanNumberStrings(s)
	s = strings.Split(s, ".")[0]
	n, ok := n.SetString(s, 10)
	if !ok {
		panic(fmt.Sprintf("unable to parse for big int with string: %s", s))
	}
	return *n
}

func cleanNumberStrings(s string) string {
	if strings.ToLower(s) == "n/a" || strings.TrimSpace(s) == "" {
		s = "0.00"
	}
	s = strings.ReplaceAll(s, "$", "")
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, "%", "")
	return s
}
