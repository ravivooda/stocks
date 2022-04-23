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
}

type Generator interface {
	Generate(ctx context.Context, analysisMap map[models.LETFAccountTicker][]models.LETFOverlapAnalysis, letfs map[models.LETFAccountTicker][]models.LETFHolding) (bool, error)
}

type generator struct {
	config Config
}

const summaryTemplateLoc = "website/letf/letf_summary.tmpl"

func (g *generator) Generate(_ context.Context, analysisMap map[models.LETFAccountTicker][]models.LETFOverlapAnalysis, letfs map[models.LETFAccountTicker][]models.LETFHolding) (bool, error) {
	_, err := utils.MakeDir(g.config.WebsiteDirectoryRoot)
	summariesFileRoot := fmt.Sprintf("%s/summary", g.config.WebsiteDirectoryRoot)
	_, err = utils.MakeDir(summariesFileRoot)
	if err != nil {
		return false, err
	}

	for LETFTicker, allAnalysis := range analysisMap {
		outputFilePath := fmt.Sprintf("%s/%s.html", summariesFileRoot, LETFTicker)
		b, err := g.logToHTML(summaryTemplateLoc, outputFilePath, LETFTicker, letfs[LETFTicker], allAnalysis)
		if err != nil {
			return b, err
		}
	}

	return true, nil
}

func (g *generator) logToHTML(
	summaryTemplateLoc string,
	outputFilePath string,
	accountTicker models.LETFAccountTicker,
	letfHoldings []models.LETFHolding,
	allAnalysis []models.LETFOverlapAnalysis,
) (bool, error) {
	t := template.Must(template.ParseFiles(summaryTemplateLoc))
	output, err := os.Create(outputFilePath)
	defer func(output *os.File) {
		_ = output.Close()
	}(output)
	if err != nil {
		return false, err
	}
	sort.Slice(allAnalysis, func(i, j int) bool {
		return allAnalysis[i].OverlapPercentage < allAnalysis[j].OverlapPercentage
	})
	err = t.Execute(output, struct {
		AccountTicker models.LETFAccountTicker
		Holdings      []models.LETFHolding
		Overlaps      []models.LETFOverlapAnalysis
	}{
		AccountTicker: accountTicker,
		Holdings:      letfHoldings,
		Overlaps:      allAnalysis,
	})
	if err != nil {
		return false, err
	}
	return false, nil
}

func New(config Config) Generator {
	return &generator{config: config}
}
