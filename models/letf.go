package models

type LETFAccountTicker string
type StockTicker string

type LETFHolding struct {
	//TODO: Rename this as Holding instead
	TradeDate         string
	LETFAccountTicker LETFAccountTicker
	LETFDescription   string
	StockTicker       StockTicker
	StockDescription  string
	Shares            int64
	Price             int64
	NotionalValue     float64
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
	LETFHolder        LETFAccountTicker
	LETFHoldees       []LETFAccountTicker
	OverlapPercentage float64
	DetailedOverlap   []LETFOverlap `json:"detailed_overlap"`
}
