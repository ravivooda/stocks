package overlap

import (
	"math"
	"stocks/models"
	"stocks/utils"
)

type Generator interface {
	Generate([]models.LETFHolding) []models.LETFOverlapAnalysis
}

type generator struct {
}

func (g *generator) Generate(holdings []models.LETFHolding) []models.LETFOverlapAnalysis {
	holdingsWithAccountTicker := utils.MapLETFHoldingsWithAccountTicker(holdings)
	if len(holdingsWithAccountTicker) <= 1 {
		return []models.LETFOverlapAnalysis{}
	}
	var outputs []models.LETFOverlapAnalysis
	for lkey, lLETFHoldings := range holdingsWithAccountTicker {
		for rkey, rLETFHoldings := range holdingsWithAccountTicker {
			if lkey != rkey {
				outputs = append(outputs, g.compare(lLETFHoldings, rLETFHoldings))
			}
		}
	}
	return outputs
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
			Percentage: percentMap.minPercentage,
			IndividualPercentagesMap: map[models.LETFAccountTicker]float64{
				l[0].LETFAccountTicker: percentMap.lLetfPercentage,
				r[0].LETFAccountTicker: percentMap.rLetfPercentage,
			},
		})
	}
	return models.LETFOverlapAnalysis{
		LETFHolding1:      l[0].LETFAccountTicker,
		LETFHolding2:      r[0].LETFAccountTicker,
		OverlapPercentage: totalOverlapPercentage,
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
			rLetfPercentage: holding.Percent,
		}
	}
	return rets
}

func NewOverlapGenerator() Generator {
	return &generator{}
}
