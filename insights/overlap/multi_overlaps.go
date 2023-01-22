package overlap

import (
	"math"
	"stocks/models"
	"stocks/utils"
)

func (g *generator) MergeInsights(
	analysis map[models.LETFAccountTicker][]models.LETFOverlapAnalysis,
	letfHoldingsMap map[models.LETFAccountTicker][]models.LETFHolding,
) map[models.LETFAccountTicker][]models.LETFOverlapAnalysis {
	var mappedMergedInsights = map[models.LETFAccountTicker][]models.LETFOverlapAnalysis{}
	for ticker, analyses := range analysis {
		combinations := utils.Combinations(analyses, 2)
		for _, combination := range combinations {
			var (
				targetedPercentageMatrix   = letfHoldingsMap[ticker]
				combinedPercentageMatrices [][]models.LETFHolding
				holdees                    []models.LETFAccountTicker
				maxPercentage              = float64(0)
				combinedPercentage         = float64(0)
			)
			for _, c := range combination {
				accountTicker := c.LETFHoldees[0]
				combinedPercentageMatrices = append(combinedPercentageMatrices, letfHoldingsMap[accountTicker])
				holdees = append(holdees, accountTicker)
				maxPercentage = math.Max(c.OverlapPercentage, maxPercentage)
				combinedPercentage += c.OverlapPercentage
			}
			if int(combinedPercentage) < g.c.MergedThreshold {
				continue
			}
			// TODO: Fix the merge insights logic with the new merge logic
			holdings, mappedPercentageHoldings := utils.MergeHoldings(combinedPercentageMatrices...)
			totalOverlapPercentage, overlapAnalysis := g.Compare(utils.MapLETFHoldingsWithStockTicker(targetedPercentageMatrix), utils.MapLETFHoldingsWithStockTicker(holdings))
			if z := int(totalOverlapPercentage); z >= g.c.MergedThreshold && z-int(maxPercentage) >= g.c.MinimumIncrementThreshold {
				var computedOverlaps []models.LETFOverlap
				for _, overlap := range overlapAnalysis {
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
					OverlapPercentage: totalOverlapPercentage,
					DetailedOverlap:   &computedOverlaps,
				})
			}
		}
	}
	return mappedMergedInsights
}
