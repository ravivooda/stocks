package overlap

import (
	"math"
	"stocks/models"
	"stocks/utils"
)

//for _, generator := range request.websiteGenerators {
//
//	//for leverage, analyses := range mappedOverlapAnalysis {
//	//	//fmt.Printf("Fetching merge insights for ticker %s, with leverage %s, len = %d\n", ticker, leverage, len(analyses))
//	//	if leverage == "" {
//	//		for _, analysis := range analyses {
//	//			panic(fmt.Sprintf("found empty leverage for %s\n", analysis.LETFHoldees))
//	//		}
//	//	}
//	//	mergedInsights := iGenerator.MergeInsights(map[models.LETFAccountTicker][]models.LETFOverlapAnalysis{ticker: analyses}, holdingsWithAccountTickerMap)
//	//	//fmt.Printf("Found %d merged insights\n", len(mergedInsights[ticker]))
//	//	analyses = append(analyses, mergedInsights[ticker]...)
//	//	mappedOverlapAnalysis[leverage] = analyses
//	//}
//	// TODO: Can we parallelize this stuff?
//	// Generate ETF summaries
//	_, err := generator.GenerateETF(ctx, ticker, mappedOverlapAnalysis, holdingsWithAccountTickerMap, holdingsWithStockTickerMap)
//	if err != nil {
//		panic(err)
//	}
//}

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
