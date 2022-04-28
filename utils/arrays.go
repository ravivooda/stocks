package utils

import "strings"

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
