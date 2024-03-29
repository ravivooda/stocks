package website

import (
	"stocks/external/stocks/alphavantage"
	"stocks/models"
)

type TemplateCustomMetadata struct {
	SideBarMetadata SideBarMetadata
	HeaderMetadata  HeaderMetadata
	FooterMetadata  FooterMetadata
	WebsitePaths    Paths
}

type SocialNetworkMetadata struct {
	LinkedInURL string
	FacebookURL string
	TwitterURL  string
}

type SideBarMetadata struct {
	TopETFs               []models.LETFAccountTicker
	TopStocks             []models.StockTicker
	SocialNetworkMetadata SocialNetworkMetadata
}

type TaxLossCalculationData struct {
	Begin         alphavantage.LinearTimeSeriesDaily
	Today         alphavantage.LinearTimeSeriesDaily
	IsHarvestable bool
	ChangePrice   string
	Swappables    []models.LETFAccountTicker
}

type ChartData struct {
	Ticker                 string
	LinearDailyData        []alphavantage.LinearTimeSeriesDaily
	TaxLossCalculationData TaxLossCalculationData
}

type HeaderMetadata struct {
}

type FooterMetadata struct {
}
