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

type MSHolding struct {
	NetChange        float64 `json:"netChange"`
	Volume           int     `json:"volume"`
	Ticker           string  `json:"ticker"`
	PerformanceID    string  `json:"performanceID"`
	Name             string  `json:"name"`
	Exchange         string  `json:"exchange"`
	PercentNetChange float64 `json:"percentNetChange"`
	LastPrice        float64 `json:"lastPrice"`
}

type MSResponse struct {
	Gainers []MSHolding `json:"gainers"`
	Actives []MSHolding `json:"actives"`
	Losers  []MSHolding `json:"losers"`
}
