package letf

import (
	"context"
	"fmt"
	"sort"
	"stocks/models"
)

func (g *generator) logSummaryToHTML(
	summaryTemplateLoc string,
	outputFilePath string,
	accountTicker models.LETFAccountTicker,
	letfHoldings []models.LETFHolding,
	allAnalysis map[string][]models.LETFOverlapAnalysis,
	letfs map[models.LETFAccountTicker][]models.LETFHolding,
) (bool, error) {
	data := struct {
		AccountTicker models.LETFAccountTicker
		Holdings      []models.LETFHolding
		Overlaps      map[string][]models.LETFOverlapAnalysis
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

func (g *generator) GenerateETF(
	_ context.Context,
	etf models.LETFAccountTicker,
	mappedAnalysisArray map[string][]models.LETFOverlapAnalysis,
	letfs map[models.LETFAccountTicker][]models.LETFHolding,
	stocksMap map[models.StockTicker][]models.LETFHolding,
) (bool, error) {
	summaryOutputFilePath := fmt.Sprintf("%s/%s.html", g.letfSummariesFileRoot, etf)
	for _, analysisArray := range mappedAnalysisArray {
		sort.Slice(analysisArray, func(i, j int) bool {
			return analysisArray[i].OverlapPercentage > analysisArray[j].OverlapPercentage
		})
	}

	// Generate Summary for the ticker
	if b, err := g.logSummaryToHTML(letfSummaryTemplateLoc, summaryOutputFilePath, etf, letfs[etf], mappedAnalysisArray, letfs); err != nil {
		return b, err
	}

	// Generate Overlap details
	//for _, analysisArray := range mappedAnalysisArray {
	//	for _, analysis := range analysisArray {
	//		if int(analysis.OverlapPercentage) < g.config.MinThreshold {
	//			continue
	//		}
	//		overlapOutputFilePath := fmt.Sprintf(
	//			"%s/%s_%s.html",
	//			g.overlapsFileRoot,
	//			analysis.LETFHolder,
	//			utils.JoinLETFAccountTicker(analysis.LETFHoldees, "_"),
	//		)
	//		b, err := g.logOverlapToHTML(overlapTemplateLoc, overlapOutputFilePath, analysis, stocksMap)
	//		if err != nil {
	//			return b, err
	//		}
	//	}
	//}
	return false, nil
}
