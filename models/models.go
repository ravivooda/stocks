package models

type OutstandingShares struct {
	LineNumber int
	Prefix     string
}

type Header struct {
	SkippableLines    int
	ExpectedColumns   []string
	OutstandingShares OutstandingShares
}

type Seed struct {
	URL    string
	Ticker string
	Header Header
}

type Holding struct {
	TradeDate     string
	AccountTicker string
	StockTicker   string
	Description   string
	Shares        int64
	Price         int64
	MarketValue   int64
	Percent       float64
}
