package utils

import "math"

func RoundedPercentage(x float64) float64 {
	return math.Round(x*100) / 100
}

func RoundedDouble(x float64) float64 {
	return math.Round(x*100) / 100
}
