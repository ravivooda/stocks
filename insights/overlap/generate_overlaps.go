package overlap

import (
	"fmt"
	"math"
	"stocks/models"
	"stocks/utils"
)

type Generator interface {
	Generate(holdingsWithAccountTickerMap map[models.LETFAccountTicker][]models.LETFHolding) map[models.LETFAccountTicker][]models.LETFOverlapAnalysis
	MergeInsights(
		analysis map[models.LETFAccountTicker][]models.LETFOverlapAnalysis,
		letfHoldingsMap map[models.LETFAccountTicker][]models.LETFHolding,
	) map[models.LETFAccountTicker][]models.LETFOverlapAnalysis
}

type generator struct {
	c Config
}

func (g *generator) Generate(holdingsWithAccountTickerMap map[models.LETFAccountTicker][]models.LETFHolding) map[models.LETFAccountTicker][]models.LETFOverlapAnalysis {
	if len(holdingsWithAccountTickerMap) <= 1 {
		return map[models.LETFAccountTicker][]models.LETFOverlapAnalysis{}
	}
	var (
		i        = 0
		skipped  = 0
		possible = len(holdingsWithAccountTickerMap) * len(holdingsWithAccountTickerMap)
		outputs  = make([]models.LETFOverlapAnalysis, 0, possible)
	)
	var visited = map[string]bool{}
	for lkey := range holdingsWithAccountTickerMap {
		for rkey := range holdingsWithAccountTickerMap {
			i += 1
			if lkey != rkey {
				if visited[fmt.Sprintf("%s_%s", lkey, rkey)] {
					skipped += 1
					continue
				}
				visited[fmt.Sprintf("%s_%s", lkey, rkey)] = true
				visited[fmt.Sprintf("%s_%s", rkey, lkey)] = true
				var (
					lLETFHoldingsMap = utils.MapLETFHoldingsWithStockTicker(holdingsWithAccountTickerMap[lkey])
					rLETFHoldingsMap = utils.MapLETFHoldingsWithStockTicker(holdingsWithAccountTickerMap[rkey])
				)

				if !utils.HasIntersection(lLETFHoldingsMap, rLETFHoldingsMap) {
					skipped += 1
					continue
				}
				totalOverlapPercentage, details := g.compare(lLETFHoldingsMap, rLETFHoldingsMap)
				if int(totalOverlapPercentage) > g.c.MinThreshold {
					outputs = append(outputs,
						models.LETFOverlapAnalysis{
							LETFHolder:        lkey,
							LETFHoldees:       []models.LETFAccountTicker{rkey},
							OverlapPercentage: totalOverlapPercentage,
							DetailedOverlap:   &details,
						},
						models.LETFOverlapAnalysis{
							LETFHolder:        rkey,
							LETFHoldees:       []models.LETFAccountTicker{lkey},
							OverlapPercentage: totalOverlapPercentage,
							DetailedOverlap:   &details,
						},
					)
				}
			}
			if i%1000 == 0 {
				fmt.Printf("working on %d, skipped %d, out of %d\n", i, skipped, possible)
			}
		}
	}

	mappedOutputs := utils.MapLETFAnalysisWithAccountTicker(outputs)
	return mappedOutputs
}

func (g *generator) compare(l map[models.StockTicker][]models.LETFHolding, r map[models.StockTicker][]models.LETFHolding) (float64, []models.LETFOverlap) {
	var totalOverlapPercentage float64 = 0
	var details []models.LETFOverlap
	var filledStocks = map[models.StockTicker]bool{}
	for stockTicker, lHoldings := range l {
		lHolding := lHoldings[0]
		if rHoldings, ok := r[stockTicker]; ok {
			rHolding := rHoldings[0]
			minPercentage := math.Min(lHolding.PercentContained, rHolding.PercentContained)
			totalOverlapPercentage += minPercentage
			details = append(details, models.LETFOverlap{
				Ticker:     stockTicker,
				Percentage: utils.RoundedPercentage(minPercentage),
				IndividualPercentagesMap: map[models.LETFAccountTicker]float64{
					lHolding.LETFAccountTicker: lHolding.PercentContained,
					rHolding.LETFAccountTicker: rHolding.PercentContained,
				},
			})
			filledStocks[stockTicker] = true
		} else {
			details = append(details, models.LETFOverlap{
				Ticker:     stockTicker,
				Percentage: 0,
				IndividualPercentagesMap: map[models.LETFAccountTicker]float64{
					lHolding.LETFAccountTicker: lHolding.PercentContained,
				},
			})
		}
	}

	for ticker, rHoldings := range r {
		if filledStocks[ticker] {
			continue
		}
		rHolding := rHoldings[0]
		details = append(details, models.LETFOverlap{
			Ticker:     ticker,
			Percentage: 0,
			IndividualPercentagesMap: map[models.LETFAccountTicker]float64{
				rHolding.LETFAccountTicker: rHolding.PercentContained,
			},
		})
	}
	return totalOverlapPercentage, details
}

type Config struct {
	MinThreshold              int `mapstructure:"min_threshold_percentage"`
	MergedThreshold           int `mapstructure:"min_merged_threshold_percentage"`
	MinimumIncrementThreshold int `mapstructure:"min_merged_improvement_threshold_percentage"`
}

func NewOverlapGenerator(config Config) Generator {
	return &generator{c: config}
}
