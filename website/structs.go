package website

import "stocks/models"

type TemplateCustomMetadata struct {
	SideBarMetadata SideBarMetadata
	HeaderMetadata  HeaderMetadata
	FooterMetadata  FooterMetadata
	WebsitePaths    Paths
}

type SideBarMetadata struct {
	TopETFs   []models.LETFAccountTicker
	TopStocks []models.StockTicker
}

type HeaderMetadata struct {
}

type FooterMetadata struct {
}
