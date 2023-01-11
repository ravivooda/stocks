package website

import "stocks/models"

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

type HeaderMetadata struct {
}

type FooterMetadata struct {
}
