package letf

import (
	"sort"
	"stocks/models"
)

func (g *generator) logOverlapToHTML(
	overlapTemplateLoc string,
	overlapOutputFilePath string,
	ptrAnalysis models.LETFOverlapAnalysis,
	letfs map[models.StockTicker][]models.LETFHolding,
) (bool, error) {
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
