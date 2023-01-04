package models

type StockAutoCompleteMetadata struct {
	//StockDescription string `json:"d"`
}

type StockMetadata struct {
	StockTicker      StockTicker
	StockDescription string
	ETFCount         int
}

type StockCombination struct {
	Tickers       []StockTicker
	ETF           ETFMetadata
	SummedPercent float64
}
