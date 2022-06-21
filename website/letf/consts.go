package letf

import "fmt"

const (
	letfSummaryTemplateLoc  = "website/letf/tmpl/letf_summary.tmpl"
	overlapTemplateLoc      = "website/letf/tmpl/letf_overlap.tmpl"
	welcomeTemplateLoc      = "website/letf/tmpl/letf_welcome.tmpl"
	stockSummaryTemplateLoc = "website/letf/tmpl/stock_summary.tmpl"
)

var (
	letfSummariesPathFromRoot  = "/letf-summary"
	stockSummariesPathFromRoot = "/stock-summary"
	overlapsPathFromRoot       = fmt.Sprintf("%s/overlap", letfSummariesPathFromRoot)

	websitePaths = WebsitePaths{
		LETFSummary:  letfSummariesPathFromRoot,
		Overlaps:     overlapsPathFromRoot,
		StockSummary: stockSummariesPathFromRoot,
	}
)
