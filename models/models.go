package models

type Seed struct {
	URL             string
	Ticker          string
	SkippableLines  int
	ExpectedColumns []string
}

type Holding struct {
	TradeDate     string
	AccountTicker string
	StockTicker   string
	Description   string
	Shares        int64
	Price         float64
	MarketValue   float64
	Percent       float64
}
