package letf

import (
	"context"
	"fmt"
	"html/template"
	"os"
	"sort"
	"stocks/models"
	"stocks/utils"
)

type Config struct {
	WebsiteDirectoryRoot string
	MinThreshold         int
}

type Request struct {
	AnalysisMap map[models.LETFAccountTicker][]models.LETFOverlapAnalysis
	Letfs       map[models.LETFAccountTicker][]models.LETFHolding
	StocksMap   map[models.StockTicker][]models.LETFHolding
}

type Generator interface {
	Generate(ctx context.Context, request Request) (bool, error)
}

type generator struct {
	config Config
}

type WebsitePaths struct {
	LETFSummary  string
	StockSummary string
	Overlaps     string
}

const (
	letfSummaryTemplateLoc  = "website/letf/letf_summary.tmpl"
	overlapTemplateLoc      = "website/letf/letf_overlap.tmpl"
	welcomeTemplateLoc      = "website/letf/letf_welcome.tmpl"
	stockSummaryTemplateLoc = "website/letf/stock_summary.tmpl"
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

func (g *generator) Generate(_ context.Context, request Request) (bool, error) {
	letfSummariesFileRoot := g.getFilePath(letfSummariesPathFromRoot)
	overlapsFileRoot := g.getFilePath(overlapsPathFromRoot)
	stockSummariesFileRoot := g.getFilePath(stockSummariesPathFromRoot)
	b, err := utils.MakeDirs([]string{g.config.WebsiteDirectoryRoot, letfSummariesFileRoot, overlapsFileRoot, stockSummariesFileRoot})
	if err != nil {
		return b, err
	}

	b, err = g.logWelcomePageToHTML(welcomeTemplateLoc, fmt.Sprintf("%s/index.html", g.config.WebsiteDirectoryRoot), request)
	if err != nil {
		return b, err
	}

	b, err = g.logWelcomePageToHTML(welcomeTemplateLoc, fmt.Sprintf("%s/404.html", g.config.WebsiteDirectoryRoot), request)
	if err != nil {
		return b, err
	}

	var i = 0
	for ticker, holdings := range request.StocksMap {
		escapedTickerString := string(ticker)
		stockSummaryFilePath := fmt.Sprintf("%s/%s.html", stockSummariesFileRoot, escapedTickerString)
		_, err = g.logStockSummaryPageToHTML(stockSummaryTemplateLoc, stockSummaryFilePath, escapedTickerString, holdings)
		if err != nil {
			return false, err
		}
		if i%10000 == 0 {
			fmt.Printf("logged stock summary page %d out of %d\n", i, len(request.StocksMap))
		}
		i += 1
	}

	i = 0
	for LETFTicker, holdings := range request.Letfs {
		summaryOutputFilePath := fmt.Sprintf("%s/%s.html", letfSummariesFileRoot, LETFTicker)
		allAnalysis := request.AnalysisMap[LETFTicker]
		sort.Slice(allAnalysis, func(i, j int) bool {
			return allAnalysis[i].OverlapPercentage > allAnalysis[j].OverlapPercentage
		})

		// Generate Summary for the ticker
		if b, err := g.logSummaryToHTML(letfSummaryTemplateLoc, summaryOutputFilePath, LETFTicker, holdings, allAnalysis, request.Letfs); err != nil {
			return b, err
		}

		// Generate Overlap details
		for _, analysis := range allAnalysis {
			if int(analysis.OverlapPercentage) < g.config.MinThreshold {
				continue
			}
			overlapOutputFilePath := fmt.Sprintf("%s/%s_%s.html", overlapsFileRoot, analysis.LETFHolder, utils.JoinLETFAccountTicker(analysis.LETFHoldees, "_"))
			b, err := g.logOverlapToHTML(overlapTemplateLoc, overlapOutputFilePath, analysis, request.StocksMap)
			if err != nil {
				return b, err
			}
		}
		if i%100 == 0 {
			fmt.Printf("logged letf summary and overlap page %d out of %d\n", i, len(request.Letfs))
		}
		i += 1
	}

	return true, nil
}

func (g *generator) getFilePath(pathFromRoot string) string {
	return fmt.Sprintf("%s%s", g.config.WebsiteDirectoryRoot, pathFromRoot)
}

func (g *generator) logOverlapToHTML(overlapTemplateLoc string, overlapOutputFilePath string, ptrAnalysis models.LETFOverlapAnalysis, letfs map[models.StockTicker][]models.LETFHolding) (bool, error) {
	type UnPtrAnalysis struct {
		LETFHolder        models.LETFAccountTicker
		LETFHoldees       []models.LETFAccountTicker
		OverlapPercentage float64
		DetailedOverlap   []models.LETFOverlap `json:"detailed_overlap"`
	}
	analysis := UnPtrAnalysis{
		LETFHolder:        ptrAnalysis.LETFHolder,
		LETFHoldees:       ptrAnalysis.LETFHoldees,
		OverlapPercentage: ptrAnalysis.OverlapPercentage,
		DetailedOverlap:   *ptrAnalysis.DetailedOverlap,
	}
	sort.Slice(analysis.DetailedOverlap, func(i, j int) bool {
		return analysis.DetailedOverlap[i].Percentage > analysis.DetailedOverlap[j].Percentage
	})
	var data = struct {
		Analysis     UnPtrAnalysis
		StocksMap    map[models.StockTicker][]models.LETFHolding
		WebsitePaths WebsitePaths
	}{
		Analysis:     analysis,
		StocksMap:    letfs,
		WebsitePaths: websitePaths,
	}
	return g.logHTMLWithData(overlapTemplateLoc, overlapOutputFilePath, data)
}

func (g *generator) logSummaryToHTML(summaryTemplateLoc string, outputFilePath string, accountTicker models.LETFAccountTicker, letfHoldings []models.LETFHolding, allAnalysis []models.LETFOverlapAnalysis, letfs map[models.LETFAccountTicker][]models.LETFHolding) (bool, error) {
	data := struct {
		AccountTicker models.LETFAccountTicker
		Holdings      []models.LETFHolding
		Overlaps      []models.LETFOverlapAnalysis
		AccountsMap   map[models.LETFAccountTicker][]models.LETFHolding
		WebsitePaths  WebsitePaths
	}{
		AccountTicker: accountTicker,
		Holdings:      letfHoldings,
		Overlaps:      allAnalysis,
		AccountsMap:   letfs,
		WebsitePaths:  websitePaths,
	}
	return g.logHTMLWithData(summaryTemplateLoc, outputFilePath, data)
}

func (g *generator) logWelcomePageToHTML(welcomePageTemplateLoc, outputFilePath string, request Request) (bool, error) {
	var mapped = map[string]map[models.LETFAccountTicker]bool{}
	for ticker, holdings := range request.Letfs {
		providerMap := mapped[holdings[0].Provider]
		if providerMap == nil {
			providerMap = map[models.LETFAccountTicker]bool{}
		}
		providerMap[ticker] = true
		mapped[holdings[0].Provider] = providerMap
	}
	var groupedStocks = map[string][]models.StockTicker{}
	for ticker := range request.StocksMap {
		s := "unknown"
		if len(ticker) > 0 {
			s = string(ticker[0:1])
		}
		a := groupedStocks[s]
		if a == nil {
			a = []models.StockTicker{}
		}
		a = append(a, ticker)
		groupedStocks[s] = a
	}
	var data = struct {
		TotalProvider int
		TotalSeeds    int
		Providers     map[string]map[models.LETFAccountTicker]bool
		Stocks        map[string][]models.StockTicker
		WebsitePaths  WebsitePaths
	}{
		TotalProvider: len(mapped),
		TotalSeeds:    len(request.Letfs),
		Providers:     mapped,
		Stocks:        groupedStocks,
		WebsitePaths:  websitePaths,
	}
	return g.logHTMLWithData(welcomePageTemplateLoc, outputFilePath, data)
}

func (g *generator) logStockSummaryPageToHTML(stockTemplateLoc string, outputFilePath string, ticker string, holdings []models.LETFHolding) (bool, error) {
	data := struct {
		Ticker       string
		Holdings     []models.LETFHolding
		WebsitePaths WebsitePaths
	}{
		Ticker:       ticker,
		Holdings:     holdings,
		WebsitePaths: websitePaths,
	}
	return g.logHTMLWithData(stockTemplateLoc, outputFilePath, data)
}

func (g *generator) logHTMLWithData(templateLoc string, outputFilePath string, data interface{}) (bool, error) {
	t := template.Must(template.ParseFiles(templateLoc))
	//outputFilePath = strings.ReplaceAll(outputFilePath, "/", "_")
	outputFile, err := os.Create(outputFilePath)
	defer func(outputFile *os.File) {
		_ = outputFile.Close()
	}(outputFile)
	if err != nil {
		return false, err
	}
	err = t.Execute(outputFile, data)
	if err != nil {
		return false, err
	}
	return false, nil
}

func New(config Config) Generator {
	return &generator{config: config}
}
