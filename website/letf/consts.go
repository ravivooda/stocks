package letf

import "fmt"

const (
	overlapTemplateLoc      = "website/letf/tmpl/letf_overlap.tmpl"
	welcomeTemplateLoc      = "website/letf/tmpl/letf_welcome.tmpl"
	stockSummaryTemplateLoc = "website/letf/tmpl/stock_summary.tmpl"
)

var (
	ETFSummaryTemplate         = "letf_summary.tmpl"
	TemplatesDir               = "website/letf/tmpl"
	ETFSummaryTemplateLoc      = fmt.Sprintf("%s/%s", TemplatesDir, ETFSummaryTemplate)
	etfSummariesPathFromRoot   = "/etf-summary"
	stockSummariesPathFromRoot = "/stock-summary"
	overlapsPathFromRoot       = fmt.Sprintf("%s/overlap", etfSummariesPathFromRoot)

	DefaultWebsitePaths = WebsitePaths{
		LETFSummary:      etfSummariesPathFromRoot,
		Overlaps:         overlapsPathFromRoot,
		StockSummary:     stockSummariesPathFromRoot,
		TemplatesRootDir: TemplatesDir,
	}
)
