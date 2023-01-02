package utils

import (
	"strconv"
	"strings"
)

func ParseInt(s string) int64 {
	s = strings.Split(s, ".")[0]
	ri, _ := strconv.ParseInt(s, 10, 64)
	return ri
}

func ParseFloat(s string) float64 {
	p, _ := strconv.ParseFloat(s, 64)
	return p
}
