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
	ETFSummaryTemplateLoc      = "website/letf/tmpl/letf_summary.tmpl"
	letfSummariesPathFromRoot  = "/letf-summary"
	stockSummariesPathFromRoot = "/stock-summary"
	overlapsPathFromRoot       = fmt.Sprintf("%s/overlap", letfSummariesPathFromRoot)

	websitePaths = WebsitePaths{
		LETFSummary:  letfSummariesPathFromRoot,
		Overlaps:     overlapsPathFromRoot,
		StockSummary: stockSummariesPathFromRoot,
	}
)
