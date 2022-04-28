package models

type LETFAccountTicker string
type StockTicker string

type LETFHolding struct {
	TradeDate         string
	LETFAccountTicker LETFAccountTicker
	LETFDescription   string
	StockTicker       StockTicker
	StockDescription  string
	Shares            int64
	Price             int64
	MarketValue       int64
	PercentContained  float64
	Provider          string
}

type LETFOverlap struct {
	Ticker                   StockTicker
	Percentage               float64
	IndividualPercentagesMap map[LETFAccountTicker]float64
}

type LETFOverlapAnalysis struct {
	LETFHolding1      LETFAccountTicker
	LETFHolding2      LETFAccountTicker
	OverlapPercentage float64
	DetailedOverlap   []LETFOverlap `json:"detailed_overlap"`
}
