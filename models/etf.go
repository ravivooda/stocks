package models

import (
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
	"time"
)

type ETF struct {
	Symbol                        LETFAccountTicker
	ETFName                       string
	AssetClass                    string
	TotalAssets                   big.Int
	YTDPriceChange                float64
	AvgDailyVolume                big.Int
	PreviousClosingPrice          big.Int
	OneDayChange                  float64
	Inverse                       bool
	Leveraged                     string
	OneWeek                       float64
	OneMonth                      float64
	OneYear                       float64
	ThreeYear                     float64
	FiveYear                      float64
	YTDFF                         big.Int
	OneWeekFF                     big.Int
	FourWeekFF                    big.Int
	OneYearFF                     big.Int
	ThreeYearFF                   big.Int
	FiveYearFF                    big.Int
	ETFDatabaseCategory           string
	Inception                     *time.Time
	ER                            float64
	CommissionFree                string
	AnnualDividendRate            float64
	DividendDate                  *time.Time
	Dividend                      float64
	AnnualDividendYieldPercentage float64
	PERatio                       float64
	Beta                          float64
	NumberOfHoldings              int
	PercentageInTop10             float64
	STCapGainRate                 float64
	LTCapGainRate                 float64
	TaxForm                       string
	LowerBollinger                float64
	UpperBollinger                float64
	Support1                      float64
	Resistance1                   float64
	RSI                           float64
	LiquidityRating               string
	ExpensesRating                string
	ReturnsRating                 string
	VolatilityRating              string
	DividendRating                string
	ConcentrationRating           string
	ESGScore                      float64
	ESGScorePeerPercentile        float64
	ESGScoreGlobalPercentile      float64
	CarbonIntensity               float64
	SRIExclusionCriteria          float64
	SustainableImpactSolutions    float64
}

func GenerateETFFromStrings(input []string) ETF {
	return ETF{
		Symbol:                        LETFAccountTicker(input[0]),
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
	if strings.ToLower(s) == "false" {
		return ""
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
