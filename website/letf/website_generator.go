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

type Generator interface {
	Generate(ctx context.Context, analysisMap map[models.LETFAccountTicker][]models.LETFOverlapAnalysis, letfs map[models.LETFAccountTicker][]models.LETFHolding, stocksMap map[models.StockTicker][]models.LETFHolding) (bool, error)
}

type generator struct {
	config Config
}

type WebsitePaths struct {
	Summary  string
	Overlaps string
}

const (
	summaryTemplateLoc = "website/letf/letf_summary.tmpl"
	overlapTemplateLoc = "website/letf/letf_overlap.tmpl"
	welcomeTemplateLoc = "website/letf/letf_welcome.tmpl"
)

var (
	summariesPathFromRoot = "/summary"
	overlapsPathFromRoot  = fmt.Sprintf("%s/overlap", summariesPathFromRoot)

	websitePaths = WebsitePaths{
		Summary:  summariesPathFromRoot,
		Overlaps: overlapsPathFromRoot,
	}
)

func (g *generator) Generate(_ context.Context, analysisMap map[models.LETFAccountTicker][]models.LETFOverlapAnalysis, letfs map[models.LETFAccountTicker][]models.LETFHolding, stocksMap map[models.StockTicker][]models.LETFHolding) (bool, error) {
	summariesFileRoot := g.getFilePath(summariesPathFromRoot)
	overlapsFileRoot := g.getFilePath(overlapsPathFromRoot)
	b, err := utils.MakeDirs([]string{g.config.WebsiteDirectoryRoot, summariesFileRoot, overlapsFileRoot})
	if err != nil {
		return b, err
	}

	b, err = g.logWelcomePageToHTML(welcomeTemplateLoc, fmt.Sprintf("%s/index.html", g.config.WebsiteDirectoryRoot), letfs)
	if err != nil {
		return b, err
	}

	b, err = g.logWelcomePageToHTML(welcomeTemplateLoc, fmt.Sprintf("%s/404.html", g.config.WebsiteDirectoryRoot), letfs)
	if err != nil {
		return b, err
	}

	for LETFTicker, holdings := range letfs {
		summaryOutputFilePath := fmt.Sprintf("%s/%s.html", summariesFileRoot, LETFTicker)
		allAnalysis := analysisMap[LETFTicker]
		sort.Slice(allAnalysis, func(i, j int) bool {
			return allAnalysis[i].OverlapPercentage > allAnalysis[j].OverlapPercentage
		})

		// Generate Summary for the ticker
		if b, err := g.logSummaryToHTML(summaryTemplateLoc, summaryOutputFilePath, LETFTicker, holdings, allAnalysis, letfs); err != nil {
			return b, err
		}

		// Generate Overlap details
		for _, analysis := range allAnalysis {
			if int(analysis.OverlapPercentage) >= g.config.MinThreshold {

			}
			overlapOutputFilePath := fmt.Sprintf("%s/%s_%s.html", overlapsFileRoot, analysis.LETFHolder, utils.JoinLETFAccountTicker(analysis.LETFHoldees, "_"))
			b, err := g.logOverlapToHTML(overlapTemplateLoc, overlapOutputFilePath, analysis, stocksMap)
			if err != nil {
				return b, err
			}
		}
	}

	return true, nil
}

func (g *generator) getFilePath(pathFromRoot string) string {
	return fmt.Sprintf("%s%s", g.config.WebsiteDirectoryRoot, pathFromRoot)
}

func (g *generator) logOverlapToHTML(overlapTemplateLoc string, overlapOutputFilePath string, analysis models.LETFOverlapAnalysis, letfs map[models.StockTicker][]models.LETFHolding) (bool, error) {
	sort.Slice(analysis.DetailedOverlap, func(i, j int) bool {
		return analysis.DetailedOverlap[i].Percentage > analysis.DetailedOverlap[j].Percentage
	})
	var data = struct {
		Analysis     models.LETFOverlapAnalysis
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

func (g *generator) logWelcomePageToHTML(welcomePageTemplateLoc, outputFilePath string, letfs map[models.LETFAccountTicker][]models.LETFHolding) (bool, error) {
	var mapped = map[string]map[models.LETFAccountTicker]bool{}
	for ticker, holdings := range letfs {
		providerMap := mapped[holdings[0].Provider]
		if providerMap == nil {
			providerMap = map[models.LETFAccountTicker]bool{}
		}
		providerMap[ticker] = true
		mapped[holdings[0].Provider] = providerMap
	}
	var data = struct {
		TotalProvider int
		TotalSeeds    int
		Providers     map[string]map[models.LETFAccountTicker]bool
		WebsitePaths  WebsitePaths
	}{
		TotalProvider: len(mapped),
		TotalSeeds:    len(letfs),
		Providers:     mapped,
		WebsitePaths:  websitePaths,
	}
	return g.logHTMLWithData(welcomePageTemplateLoc, outputFilePath, data)
}

func (g *generator) logHTMLWithData(templateLoc string, outputFilePath string, data interface{}) (bool, error) {
	t := template.Must(template.ParseFiles(templateLoc))
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
