package letf

import (
	"context"
	"fmt"
	"sort"
	"stocks/models"
	"stocks/utils"
)

func (g *generator) logSummaryToHTML(
	summaryTemplateLoc string,
	outputFilePath string,
	accountTicker models.LETFAccountTicker,
	letfHoldings []models.LETFHolding,
	allAnalysis []models.LETFOverlapAnalysis,
	letfs map[models.LETFAccountTicker][]models.LETFHolding,
) (bool, error) {
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

func (g *generator) GenerateETF(ctx context.Context, etf models.LETFAccountTicker, analysisArray []models.LETFOverlapAnalysis, letfs map[models.LETFAccountTicker][]models.LETFHolding, stocksMap map[models.StockTicker][]models.LETFHolding) (bool, error) {
	summaryOutputFilePath := fmt.Sprintf("%s/%s.html", g.letfSummariesFileRoot, etf)
	sort.Slice(analysisArray, func(i, j int) bool {
		return analysisArray[i].OverlapPercentage > analysisArray[j].OverlapPercentage
	})

	// Generate Summary for the ticker
	if b, err := g.logSummaryToHTML(letfSummaryTemplateLoc, summaryOutputFilePath, etf, letfs[etf], analysisArray, letfs); err != nil {
		return b, err
	}

	// Generate Overlap details
	for _, analysis := range analysisArray {
		if int(analysis.OverlapPercentage) < g.config.MinThreshold {
			continue
		}
		overlapOutputFilePath := fmt.Sprintf(
			"%s/%s_%s.html",
			g.overlapsFileRoot,
			analysis.LETFHolder,
			utils.JoinLETFAccountTicker(analysis.LETFHoldees, "_"),
		)
		b, err := g.logOverlapToHTML(overlapTemplateLoc, overlapOutputFilePath, analysis, stocksMap)
		if err != nil {
			return b, err
		}
	}
	return false, nil
}
