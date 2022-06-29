package models

import (
	"math/big"
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
