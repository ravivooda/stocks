package letf

import "fmt"

const (
	welcomeTemplateLoc = "website/letf/tmpl/letf_welcome.tmpl"
)

var (
	ETFSummaryTemplate         = "letf_summary.tmpl"
	TemplatesDir               = "website/letf/tmpl"
	ETFSummaryTemplateLoc      = fmt.Sprintf("%s/%s", TemplatesDir, ETFSummaryTemplate)
	etfSummariesPathFromRoot   = "/etf-summary"
	stockSummariesPathFromRoot = "/stock-summary"
	overlapsPathFromRoot       = fmt.Sprintf("%s/overlap", etfSummariesPathFromRoot)

	StockSummaryTemplate    = "stock_summary.tmpl"
	stockSummaryTemplateLoc = fmt.Sprintf("%s/%s", TemplatesDir, StockSummaryTemplate)

	OverlapTemplate    = "letf_overlap.tmpl"
	overlapTemplateLoc = fmt.Sprintf("%s/%s", TemplatesDir, OverlapTemplate)

	DefaultWebsitePaths = WebsitePaths{
		LETFSummary:      etfSummariesPathFromRoot,
		Overlaps:         overlapsPathFromRoot,
		StockSummary:     stockSummariesPathFromRoot,
		TemplatesRootDir: TemplatesDir,
	}
)
