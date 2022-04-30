package utils

import (
	"stocks/models"
	"strings"
)

func Trimmed(input []string) []string {
	var rets []string
	for _, s := range input {
		rets = append(rets, strings.TrimSpace(s))
	}
	return rets
}

func FilterNonStockRows(rows [][]string, validator func(row []string) bool) [][]string {
	var retRows [][]string
	for _, row := range rows {
		if validator(row) {
			retRows = append(retRows, row)
		}
	}
	return retRows
}

func SumHoldings(holdings []models.LETFHolding) float64 {
	var totalPercent float64
	for _, holding := range holdings {
		totalPercent += holding.PercentContained
	}
	return totalPercent
}

func JoinLETFAccountTicker(input []models.LETFAccountTicker, separator string) string {
	var rets []string
	for _, ticker := range input {
		rets = append(rets, string(ticker))
	}
	return strings.Join(rets, separator)
}
