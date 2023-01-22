package website

import (
	"fmt"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"stocks/models"
	"stocks/utils"
	"time"
)

const stockParamKey = "stock"
const (
	overlapKeyLHS = "lhs"
	overlapKeyRHS = "rhs"
)

var DefaultWebsitePaths = Paths{
	LETFSummary:      etfSummariesPathFromRoot,
	Overlaps:         overlapsPathFromRoot,
	StockSummary:     stockSummariesPathFromRoot,
	TemplatesRootDir: TemplatesDir,
}

var (
	ETFSummaryTemplate = "etf_summary.tmpl"
	TemplatesDir       = "website/template/code"

	etfSummariesPathFromRoot   = "/etf-summary"
	stockSummariesPathFromRoot = "/stock-summary"
	overlapsPathFromRoot       = fmt.Sprintf("%s/overlap", etfSummariesPathFromRoot)

	StockSummaryTemplate = "stock_summary.tmpl"

	OverlapTemplate = "etf_overlap.tmpl"

	listAllETFsTemplate   = "list_all_etfs.tmpl"
	listAllStocksTemplate = "list_all_stocks.tmpl"

	useCasesTemplate          = "use_cases.tmpl"
	disclaimerTemplate        = "disclaimer.tmpl"
	findOverlapsInputTemplate = "find_overlaps_input.tmpl"
	faqsTemplate              = "faq.tmpl"
	contactTemplate           = "contact.tmpl"

	error404tmpl  = "page-error-404.tmpl"
	themePath     = "./website/template/static/quixlab/theme"
	staticPath    = "./website/template/static"
	generatedPath = "./website/template/generated"
)

func renderETFsArray(input []models.LETFAccountTicker) string {
	return utils.JoinLETFAccountTicker(input, ",")
}

func renderPercentage(input float64) string {
	if input < 10 {
		return fmt.Sprintf("0%.2f", input)
	}
	return fmt.Sprintf("%.2f", input)
}

func renderLargeNumbers(input int) string {
	p := message.NewPrinter(language.English)
	return p.Sprintf("%d\n", input)
}

func renderStockTickersCount(input []models.StockTicker) string {
	if len(input) > 1 {
		return fmt.Sprintf("%d stocks", len(input))
	}
	return "1 stock"
}

func renderDate(input string) string {
	date, _ := time.Parse("2006-01-02", input) // TODO: Silently ignoring error when parsing date
	return date.Format("January 2, 2006")
}
