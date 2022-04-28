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

const (
	summaryTemplateLoc = "website/letf/letf_summary.tmpl"
	overlapTemplateLoc = "website/letf/letf_overlap.tmpl"
)

func (g *generator) Generate(_ context.Context, analysisMap map[models.LETFAccountTicker][]models.LETFOverlapAnalysis, letfs map[models.LETFAccountTicker][]models.LETFHolding, stocksMap map[models.StockTicker][]models.LETFHolding) (bool, error) {
	summariesFileRoot := fmt.Sprintf("%s/summary", g.config.WebsiteDirectoryRoot)
	overlapsFileRoot := fmt.Sprintf("%s/overlap", summariesFileRoot)
	_, err := utils.MakeDirs([]string{g.config.WebsiteDirectoryRoot, summariesFileRoot, overlapsFileRoot})
	if err != nil {
		return false, err
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
			overlapOutputFilePath := fmt.Sprintf("%s/%s_%s.html", overlapsFileRoot, analysis.LETFHolding1, analysis.LETFHolding2)
			b, err := g.logOverlapToHTML(overlapOutputFilePath, analysis, stocksMap)
			if err != nil {
				return b, err
			}
		}
	}

	return true, nil
}

func (g *generator) logOverlapToHTML(overlapOutputFilePath string, analysis models.LETFOverlapAnalysis, letfs map[models.StockTicker][]models.LETFHolding) (bool, error) {
	t := template.Must(template.ParseFiles(overlapTemplateLoc))
	outputFile, err := os.Create(overlapOutputFilePath)
	defer func(outputFile *os.File) {
		_ = outputFile.Close()
	}(outputFile)
	if err != nil {
		return false, err
	}
	err = t.Execute(outputFile, struct {
		Analysis  models.LETFOverlapAnalysis
		StocksMap map[models.StockTicker][]models.LETFHolding
	}{
		Analysis:  analysis,
		StocksMap: letfs,
	})
	if err != nil {
		return false, err
	}
	return false, nil
}

func (g *generator) logSummaryToHTML(summaryTemplateLoc string, outputFilePath string, accountTicker models.LETFAccountTicker, letfHoldings []models.LETFHolding, allAnalysis []models.LETFOverlapAnalysis, letfs map[models.LETFAccountTicker][]models.LETFHolding) (bool, error) {
	t := template.Must(template.ParseFiles(summaryTemplateLoc))
	output, err := os.Create(outputFilePath)
	defer func(output *os.File) {
		_ = output.Close()
	}(output)
	if err != nil {
		return false, err
	}
	err = t.Execute(output, struct {
		AccountTicker models.LETFAccountTicker
		Holdings      []models.LETFHolding
		Overlaps      []models.LETFOverlapAnalysis
		AccountsMap   map[models.LETFAccountTicker][]models.LETFHolding
	}{
		AccountTicker: accountTicker,
		Holdings:      letfHoldings,
		Overlaps:      allAnalysis,
		AccountsMap:   letfs,
	})
	if err != nil {
		return false, err
	}
	return false, nil
}

func New(config Config) Generator {
	return &generator{config: config}
}
