package letf

import (
	"context"
	"fmt"
	"html/template"
	"os"
	"stocks/models"
	"stocks/utils"
)

type Config struct {
	WebsiteDirectoryRoot string
	MinThreshold         int
}

type Request struct {
	//AnalysisMap map[models.LETFAccountTicker][]models.LETFOverlapAnalysis
	Letfs     map[models.LETFAccountTicker][]models.LETFHolding
	StocksMap map[models.StockTicker][]models.LETFHolding
	//MappedAnalysisArray map[string][]models.LETFOverlapAnalysis
}

type Generator interface {
	Generate(ctx context.Context, request Request) (bool, error)
	GenerateETF(
		ctx context.Context,
		etf models.LETFAccountTicker,
		mappedAnalysisArray map[string][]models.LETFOverlapAnalysis,
		letfs map[models.LETFAccountTicker][]models.LETFHolding,
		stocksMap map[models.StockTicker][]models.LETFHolding,
	) (bool, error)
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
	//var i = 0
	//for ticker, holdings := range request.StocksMap {
	//	escapedTickerString := string(ticker)
	//	stockSummaryFilePath := fmt.Sprintf("%s/%s.html", g.stockSummariesFileRoot, escapedTickerString)
	//	_, err = g.logStockSummaryPageToHTML(stockSummaryTemplateLoc, stockSummaryFilePath, escapedTickerString, holdings)
	//	if err != nil {
	//		return false, err
	//	}
	//	if i%10000 == 0 {
	//		fmt.Printf("logged stock summary page %d out of %d\n", i, len(request.StocksMap))
	//	}
	//	i += 1
	//}

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
