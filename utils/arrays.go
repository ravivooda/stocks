package utils

import (
	"fmt"
	"sort"
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

func MergeHoldings(holdingsList ...[]models.LETFHolding) ([]models.LETFHolding, map[models.StockTicker]map[models.LETFAccountTicker]float64) {
	var mappedHoldings = map[models.StockTicker]models.LETFHolding{}
	var originalMapping = map[models.StockTicker]map[models.LETFAccountTicker]float64{}
	for _, holdings := range holdingsList {
		for _, holding := range holdings {
			accountMap := originalMapping[holding.StockTicker]
			if accountMap == nil {
				accountMap = map[models.LETFAccountTicker]float64{}
			}
			accountMap[holding.LETFAccountTicker] = holding.PercentContained
			originalMapping[holding.StockTicker] = accountMap
			if letfHolding, ok := mappedHoldings[holding.StockTicker]; ok {
				holding.PercentContained += letfHolding.PercentContained
			}
			mappedHoldings[holding.StockTicker] = holding
		}
	}
	var mergedHoldings []models.LETFHolding
	for stock, holding := range mappedHoldings {
		if holding.PercentContained > 0 {
			mergedHoldings = append(mergedHoldings, holding)
		} else {
			delete(mappedHoldings, stock)
			delete(originalMapping, stock)
		}
	}
	return mergedHoldings, originalMapping
}

// Combinations returns combinations of n elements for a given array.
// For n < 1, it equals to All and returns all combinations.
func Combinations(set []models.LETFOverlapAnalysis, r int) (rt [][]models.LETFOverlapAnalysis) {
	pool := set
	n := len(pool)

	if r > n {
		return
	}

	indices := make([]int, r)
	for i := range indices {
		indices[i] = i
	}

	result := make([]models.LETFOverlapAnalysis, r)
	for i, el := range indices {
		result[i] = pool[el]
	}
	s2 := make([]models.LETFOverlapAnalysis, r)
	copy(s2, result)
	rt = append(rt, s2)

	for {
		i := r - 1
		for ; i >= 0 && indices[i] == i+n-r; i -= 1 {
		}

		if i < 0 {
			return
		}

		indices[i] += 1
		for j := i + 1; j < r; j += 1 {
			indices[j] = indices[j-1] + 1
		}

		for ; i < len(indices); i += 1 {
			result[i] = pool[indices[i]]
		}
		s2 = make([]models.LETFOverlapAnalysis, r)
		copy(s2, result)
		rt = append(rt, s2)
	}
}

func FilteredForPrinting(holdings []models.LETFHolding) [][]string {
	var filteredHoldings [][]string
	for _, holding := range holdings {
		filteredHoldings = append(filteredHoldings, []string{string(holding.StockTicker), fmt.Sprintf("%f", holding.PercentContained), fmt.Sprintf("%d", holding.MarketValue)})
	}
	sort.Slice(filteredHoldings, func(i, j int) bool {
		return filteredHoldings[i][1] > filteredHoldings[j][1]
	})
	return filteredHoldings
}

func HasIntersection(l map[models.StockTicker][]models.LETFHolding, r map[models.StockTicker][]models.LETFHolding) bool {
	for ticker := range r {
		if _, ok := l[ticker]; ok {
			return true
		}
	}

	return false
}

func SortOverlapsWithinLeverage(parsedOverlaps map[string][]models.LETFOverlapAnalysis) map[string][]models.LETFOverlapAnalysis {
	var sortedOverlaps = map[string][]models.LETFOverlapAnalysis{}
	for s, analyses := range parsedOverlaps {
		sort.Slice(analyses, func(i, j int) bool {
			return analyses[i].OverlapPercentage > analyses[j].OverlapPercentage
		})
		sortedOverlaps[s] = analyses
	}
	return sortedOverlaps
}

func MapToArrayForStockTickers(input map[models.StockTicker]models.StockMetadata) []models.StockTicker {
	var rets []models.StockTicker
	for ticker := range input {
		rets = append(rets, ticker)
	}
	return rets
}

func MapToArrayForAccountTickers(input map[models.LETFAccountTicker]models.ETFMetadata) []models.LETFAccountTicker {
	var rets []models.LETFAccountTicker
	for ticker := range input {
		rets = append(rets, ticker)
	}
	return rets
}
