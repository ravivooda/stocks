package models

type LETFAccountTicker string
type StockTicker string

type LETFHolding struct {
	//TODO: Rename this as Holding instead
	TradeDate         string            `json:"t,omitempty"`
	LETFAccountTicker LETFAccountTicker `json:"lat,omitempty"`
	LETFDescription   string            `json:"d,omitempty"`
	StockTicker       StockTicker       `json:"st,omitempty"`
	StockDescription  string            `json:"sd,omitempty"`
	Shares            int64             `json:"s,omitempty"`
	Price             int64             `json:"p,omitempty"`
	NotionalValue     float64           `json:"nv,omitempty"`
	MarketValue       int64             `json:"mv,omitempty"`
	PercentContained  float64           `json:"pc,omitempty"`
	Provider          string            `json:"pr,omitempty"`
}

type LETFOverlap struct {
	Ticker                   StockTicker                   `json:"t"`
	Percentage               float64                       `json:"p"`
	IndividualPercentagesMap map[LETFAccountTicker]float64 `json:"i"`
}

type LETFOverlapAnalysis struct {
	LETFHolder        LETFAccountTicker   `json:"lhr"`
	LETFHoldees       []LETFAccountTicker `json:"lhs"`
	OverlapPercentage float64             `json:"o"`
	DetailedOverlap   *[]LETFOverlap      `json:"d"`
}
