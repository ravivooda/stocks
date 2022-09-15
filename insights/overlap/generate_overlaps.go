package overlap

import (
	"fmt"
	"math"
	"stocks/models"
	"stocks/utils"
)

type Generator interface {
	Generate(holdingsWithAccountTickerMap map[models.LETFAccountTicker][]models.LETFHolding, iterator func(value []models.LETFOverlapAnalysis))
	MergeInsights(
		analysis map[models.LETFAccountTicker][]models.LETFOverlapAnalysis,
		letfHoldingsMap map[models.LETFAccountTicker][]models.LETFHolding,
	) map[models.LETFAccountTicker][]models.LETFOverlapAnalysis
	Compare(l map[models.StockTicker][]models.LETFHolding, r map[models.StockTicker][]models.LETFHolding) (float64, []models.LETFOverlap)
}

type generator struct {
	c Config
}

func (g *generator) Generate(holdingsWithAccountTickerMap map[models.LETFAccountTicker][]models.LETFHolding, iterator func(value []models.LETFOverlapAnalysis)) {
	if len(holdingsWithAccountTickerMap) <= 1 {
		return
	}
	var (
		i        = 0
		skipped  = 0
		possible = len(holdingsWithAccountTickerMap) * len(holdingsWithAccountTickerMap)
	)
	for lkey := range holdingsWithAccountTickerMap {
		var outputs []models.LETFOverlapAnalysis
		for rkey := range holdingsWithAccountTickerMap {
			i += 1
			if lkey != rkey {
				var (
					lLETFHoldingsMap = utils.MapLETFHoldingsWithStockTicker(holdingsWithAccountTickerMap[lkey])
					rLETFHoldingsMap = utils.MapLETFHoldingsWithStockTicker(holdingsWithAccountTickerMap[rkey])
				)

				if !utils.HasIntersection(lLETFHoldingsMap, rLETFHoldingsMap) {
					skipped += 1
					continue
				}
				totalOverlapPercentage, details := g.Compare(lLETFHoldingsMap, rLETFHoldingsMap)
				if int(totalOverlapPercentage) > g.c.MinThreshold {
					outputs = append(outputs,
						models.LETFOverlapAnalysis{
							LETFHolder:        lkey,
							LETFHoldees:       []models.LETFAccountTicker{rkey},
							OverlapPercentage: totalOverlapPercentage,
							DetailedOverlap:   &details,
						},
					)
				}
			}
			if i%1000 == 0 {
				fmt.Printf("working on %d, skipped %d, out of %d\n", i, skipped, possible)
				utils.PrintMemUsage()
			}
		}
		iterator(outputs)
	}
}

func (g *generator) Compare(l map[models.StockTicker][]models.LETFHolding, r map[models.StockTicker][]models.LETFHolding) (float64, []models.LETFOverlap) {
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
		}
		// By commenting the following code, we are hiding 0 overlap holdings
		//else {
		//	details = append(details, models.LETFOverlap{
		//		Ticker:     stockTicker,
		//		Percentage: 0,
		//		IndividualPercentagesMap: map[models.LETFAccountTicker]float64{
		//			lHolding.LETFAccountTicker: lHolding.PercentContained,
		//		},
		//	})
		//}
	}

	// By commenting the following code, we are hiding 0 overlap holdings
	//for ticker, rHoldings := range r {
	//	if filledStocks[ticker] {
	//		continue
	//	}
	//	rHolding := rHoldings[0]
	//	details = append(details, models.LETFOverlap{
	//		Ticker:     ticker,
	//		Percentage: 0,
	//		IndividualPercentagesMap: map[models.LETFAccountTicker]float64{
	//			rHolding.LETFAccountTicker: rHolding.PercentContained,
	//		},
	//	})
	//}
	return utils.RoundedPercentage(totalOverlapPercentage), details
}

type Config struct {
	MinThreshold              int `mapstructure:"min_threshold_percentage"`
	MergedThreshold           int `mapstructure:"min_merged_threshold_percentage"`
	MinimumIncrementThreshold int `mapstructure:"min_merged_improvement_threshold_percentage"`
}

func NewOverlapGenerator(config Config) Generator {
	return &generator{c: config}
}
