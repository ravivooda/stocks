package utils

import (
	"stocks/models"
	"strings"
)

var knownAliases = map[string]string{
	"FB":   "META",
	"BLOC": "SQ",
}

func cleanAndMap(input string) string {
	input = strings.ReplaceAll(input, "/", "_")
	trimmedInput := strings.ToUpper(strings.TrimSpace(input))
	if aliasStockName, ok := knownAliases[trimmedInput]; ok {
		trimmedInput = aliasStockName
	}
	// TODO: Handle / in ticker in a better way than the hack below
	return strings.ReplaceAll(trimmedInput, "/", "_")
}

func FetchStockTicker(input string) models.StockTicker {
	return models.StockTicker(cleanAndMap(input))
}

func FetchAccountTicker(input string) models.LETFAccountTicker {
	return models.LETFAccountTicker(cleanAndMap(input))
}

func MapLETFHoldingsWithAccountTicker(input []models.LETFHolding) map[models.LETFAccountTicker][]models.LETFHolding {
	var rets = map[models.LETFAccountTicker][]models.LETFHolding{}
	for _, holding := range input {
		var s []models.LETFHolding
		if f, ok := rets[holding.LETFAccountTicker]; ok {
			s = f
		}

		s = append(s, holding)
		rets[holding.LETFAccountTicker] = s
	}
	return rets
}

func MapLETFHoldingsWithStockTicker(holdings []models.LETFHolding) map[models.StockTicker][]models.LETFHolding {
	holdingsMap := make(map[models.StockTicker][]models.LETFHolding)
	for _, holding := range holdings {
		var holdingsArray = holdingsMap[holding.StockTicker]
		if holdingsArray == nil {
			holdingsArray = []models.LETFHolding{}
		}
		holdingsArray = append(holdingsArray, holding)
		holdingsMap[holding.StockTicker] = holdingsArray
	}
	return holdingsMap
}

func MapLETFAnalysisWithAccountTicker(analysis []models.LETFOverlapAnalysis) map[models.LETFAccountTicker][]models.LETFOverlapAnalysis {
	analysisMap := map[models.LETFAccountTicker][]models.LETFOverlapAnalysis{}
	for _, overlapAnalysis := range analysis {
		var arr []models.LETFOverlapAnalysis
		if elem, ok := analysisMap[overlapAnalysis.LETFHolder]; ok {
			arr = elem
		}
		arr = append(arr, overlapAnalysis)
		analysisMap[overlapAnalysis.LETFHolder] = arr
	}
	return analysisMap
}
