package overlap

import (
	"stocks/models"
	"stocks/utils"
)

func (g *generator) MergeInsights(
	analysis map[models.LETFAccountTicker][]models.LETFOverlapAnalysis,
	letfHoldingsMap map[models.LETFAccountTicker][]models.LETFHolding,
) map[models.LETFAccountTicker][]models.LETFOverlapAnalysis {
	var mappedMergedInsights = map[models.LETFAccountTicker][]models.LETFOverlapAnalysis{}
	for ticker, analyses := range analysis {
		for _, combination := range utils.Combinations(analyses, 2) {
			var targetedPercentageMatrix = letfHoldingsMap[ticker]
			var combinedPercentageMatrices [][]models.LETFHolding
			var holdees []models.LETFAccountTicker
			for _, c := range combination {
				accountTicker := c.LETFHoldees[0]
				combinedPercentageMatrices = append(combinedPercentageMatrices, letfHoldingsMap[accountTicker])
				holdees = append(holdees, accountTicker)
			}
			holdings, mappedPercentageHoldings := utils.MergeHoldings(combinedPercentageMatrices...)
			overlapAnalysis := g.compare(targetedPercentageMatrix, holdings)
			// TODO: Improvement should have minimum threshold as well
			if int(overlapAnalysis.OverlapPercentage) >= g.c.MergedThreshold {
				var computedOverlaps []models.LETFOverlap
				for _, overlap := range overlapAnalysis.DetailedOverlap {
					overlap.IndividualPercentagesMap = mappedPercentageHoldings[overlap.Ticker]
					computedOverlaps = append(computedOverlaps, overlap)
				}
				for _, holding := range targetedPercentageMatrix {
					m := mappedPercentageHoldings[holding.StockTicker]
					if m == nil {
						m = map[models.LETFAccountTicker]float64{}
					}
					m[ticker] = holding.PercentContained
					mappedPercentageHoldings[holding.StockTicker] = m
				}
				mappedMergedInsights[ticker] = append(mappedMergedInsights[ticker], models.LETFOverlapAnalysis{
					LETFHolder:        ticker,
					LETFHoldees:       holdees,
					OverlapPercentage: overlapAnalysis.OverlapPercentage,
					DetailedOverlap:   computedOverlaps,
				})
			}
		}
	}
	return mappedMergedInsights
}
