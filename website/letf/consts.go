package letf

import "fmt"

var (
	ETFSummaryTemplate         = "etf_summary.tmpl"
	TemplatesDir               = "website/letf/tmpl"
	ETFSummaryTemplateLoc      = fmt.Sprintf("%s/%s", TemplatesDir, ETFSummaryTemplate)
	etfSummariesPathFromRoot   = "/etf-summary"
	stockSummariesPathFromRoot = "/stock-summary"
	overlapsPathFromRoot       = fmt.Sprintf("%s/overlap", etfSummariesPathFromRoot)

	StockSummaryTemplate    = "stock_summary.tmpl"
	stockSummaryTemplateLoc = fmt.Sprintf("%s/%s", TemplatesDir, StockSummaryTemplate)

	OverlapTemplate    = "etf_overlap.tmpl"
	overlapTemplateLoc = fmt.Sprintf("%s/%s", TemplatesDir, OverlapTemplate)

	WelcomeTemplate    = "welcome.tmpl"
	welcomeTemplateLoc = fmt.Sprintf("%s/%s", TemplatesDir, WelcomeTemplate)

	DefaultWebsitePaths = WebsitePaths{
		LETFSummary:      etfSummariesPathFromRoot,
		Overlaps:         overlapsPathFromRoot,
		StockSummary:     stockSummariesPathFromRoot,
		TemplatesRootDir: TemplatesDir,
	}
)
