package models

type OutstandingShares struct {
	LineNumber int    `mapstructure:"line_number"`
	Prefix     string `mapstructure:"prefix"`
}

type Header struct {
	SkippableLines    int               `mapstructure:"skippable_lines"`
	ExpectedColumns   []string          `mapstructure:"expected_columns"`
	OutstandingShares OutstandingShares `mapstructure:"outstanding_shares"`
}

type Provider string

const (
	Direxion          = "Direxion"
	MicroSector       = "microsector"
	ProShares         = "proshares"
	Invesco           = "Invesco"
	MasterDataReports = "masterdatareports"
)

type Seed struct {
	URL      string
	Ticker   string
	Header   Header
	Provider Provider
}

type MSHolding struct {
	NetChange        TwoRoundedFloat `json:"netChange"`
	Volume           int             `json:"volume"`
	Ticker           string          `json:"ticker"`
	PerformanceID    string          `json:"performanceID"`
	Name             string          `json:"name"`
	Exchange         string          `json:"exchange"`
	PercentNetChange TwoRoundedFloat `json:"percentNetChange"`
	LastPrice        TwoRoundedFloat `json:"lastPrice"`
}

type MSResponse struct {
	Gainers []MSHolding `json:"gainers"`
	Actives []MSHolding `json:"actives"`
	Losers  []MSHolding `json:"losers"`
}
