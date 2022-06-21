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
	GenerateETF(ctx context.Context, etf models.LETFAccountTicker, analysisArray []models.LETFOverlapAnalysis, letfs map[models.LETFAccountTicker][]models.LETFHolding, stocksMap map[models.StockTicker][]models.LETFHolding) (bool, error)
}

type generator struct {
	config                 Config
	letfSummariesFileRoot  string
	overlapsFileRoot       string
	stockSummariesFileRoot string
}

type WebsitePaths struct {
	LETFSummary  string
	StockSummary string
	Overlaps     string
}

func (g *generator) Generate(_ context.Context, request Request) (bool, error) {
	b, err := g.logWelcomePageToHTML(welcomeTemplateLoc, fmt.Sprintf("%s/index.html", g.config.WebsiteDirectoryRoot), request)
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
		stockSummaryFilePath := fmt.Sprintf("%s/%s.html", g.stockSummariesFileRoot, escapedTickerString)
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
		summaryOutputFilePath := fmt.Sprintf("%s/%s.html", g.letfSummariesFileRoot, LETFTicker)
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
			overlapOutputFilePath := fmt.Sprintf("%s/%s_%s.html", g.overlapsFileRoot, analysis.LETFHolder, utils.JoinLETFAccountTicker(analysis.LETFHoldees, "_"))
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
	// TODO: Need to filter out about for

	return true, nil
}

func getFilePath(websiteDirectoryRoot string, pathFromRoot string) string {
	return fmt.Sprintf("%s%s", websiteDirectoryRoot, pathFromRoot)
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

func New(config Config) (Generator, error) {
	letfSummariesFileRoot := getFilePath(config.WebsiteDirectoryRoot, letfSummariesPathFromRoot)
	overlapsFileRoot := getFilePath(config.WebsiteDirectoryRoot, overlapsPathFromRoot)
	stockSummariesFileRoot := getFilePath(config.WebsiteDirectoryRoot, stockSummariesPathFromRoot)
	_, err := utils.MakeDirs([]string{config.WebsiteDirectoryRoot, letfSummariesFileRoot, overlapsFileRoot, stockSummariesFileRoot})
	if err != nil {
		return nil, err
	}

	return &generator{
		config:                 config,
		letfSummariesFileRoot:  letfSummariesFileRoot,
		overlapsFileRoot:       overlapsFileRoot,
		stockSummariesFileRoot: stockSummariesFileRoot,
	}, nil
}
