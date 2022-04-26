package utils

import "strings"

func Trimmed(input []string) []string {
	var rets []string
	for _, s := range input {
		rets = append(rets, strings.TrimSpace(s))
	}
	return rets
}
