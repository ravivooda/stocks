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
	trimmedInput := strings.ToUpper(strings.TrimSpace(input))
	if aliasStockName, ok := knownAliases[trimmedInput]; ok {
		trimmedInput = aliasStockName
	}
	return trimmedInput
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

func MapLETFHoldingsWithStockTicker(holdings []models.LETFHolding) map[models.StockTicker]models.LETFHolding {
	holdingsMap := make(map[models.StockTicker]models.LETFHolding)
	for _, holding := range holdings {
		holdingsMap[holding.StockTicker] = holding
	}
	return holdingsMap
}
