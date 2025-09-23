package utils

import "math"

// RoundFloat rounds a float64 to the specified number of decimal places
func RoundFloat(val float64, precision int) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}
