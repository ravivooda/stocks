package overlap

import (
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
	var outputs []models.LETFOverlapAnalysis
	for lkey, lLETFHoldings := range holdingsWithAccountTickerMap {
		for rkey, rLETFHoldings := range holdingsWithAccountTickerMap {
			if lkey != rkey {
				overlapAnalysis := g.compare(lLETFHoldings, rLETFHoldings)
				if int(overlapAnalysis.OverlapPercentage) > g.c.MinThreshold {
					outputs = append(outputs, overlapAnalysis)
				}
			}
		}
	}

	mappedOutputs := utils.MapLETFAnalysisWithAccountTicker(outputs)
	return mappedOutputs
}

func (g *generator) compare(l []models.LETFHolding, r []models.LETFHolding) models.LETFOverlapAnalysis {
	lmap := mapStockHoldings(l)
	rmap := mapStockHoldings(r)
	var totalOverlapPercentage float64 = 0
	for lstock, lpercentageMap := range lmap {
		if rpercentageMap, ok := rmap[lstock]; ok {
			minPercentage := math.Min(lpercentageMap.rLetfPercentage, rpercentageMap.rLetfPercentage)
			rmap[lstock] = holdingRowMap{
				lLetfPercentage: lpercentageMap.rLetfPercentage,
				rLetfPercentage: rpercentageMap.rLetfPercentage,
				minPercentage:   minPercentage,
			}
			totalOverlapPercentage += minPercentage
		} else {
			rmap[lstock] = holdingRowMap{
				lLetfPercentage: lpercentageMap.rLetfPercentage,
			}
		}
	}

	var details []models.LETFOverlap
	for ticker, percentMap := range rmap {
		details = append(details, models.LETFOverlap{
			Ticker:     ticker,
			Percentage: utils.RoundedPercentage(percentMap.minPercentage),
			IndividualPercentagesMap: map[models.LETFAccountTicker]float64{
				l[0].LETFAccountTicker: percentMap.lLetfPercentage,
				r[0].LETFAccountTicker: percentMap.rLetfPercentage,
			},
		})
	}
	return models.LETFOverlapAnalysis{
		LETFHolder:        l[0].LETFAccountTicker,
		LETFHoldees:       []models.LETFAccountTicker{r[0].LETFAccountTicker},
		OverlapPercentage: utils.RoundedPercentage(totalOverlapPercentage),
		DetailedOverlap:   details,
	}
}

type holdingRowMap struct {
	lLetfPercentage float64
	rLetfPercentage float64
	minPercentage   float64
}

func mapStockHoldings(h []models.LETFHolding) map[models.StockTicker]holdingRowMap {
	var rets = map[models.StockTicker]holdingRowMap{}
	for _, holding := range h {
		// TODO: Replace holdingRowMap hack
		// We are hacking rLetfPercentage to get the minimum. This really is bad.
		rets[holding.StockTicker] = holdingRowMap{
			rLetfPercentage: holding.PercentContained,
		}
	}
	return rets
}

type Config struct {
	MinThreshold              int `mapstructure:"min_threshold_percentage"`
	MergedThreshold           int `mapstructure:"min_merged_threshold_percentage"`
	MinimumIncrementThreshold int `mapstructure:"min_merged_improvement_threshold_percentage"`
}

func NewOverlapGenerator(config Config) Generator {
	return &generator{c: config}
}
