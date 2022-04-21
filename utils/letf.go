package utils

import "stocks/models"

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
