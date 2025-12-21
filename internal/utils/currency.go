package utils

import (
	"fmt"
	"strconv"
	"strings"
)

// FormatEuro formats a float as Euro currency
// Example: 100.5 -> "€100,50"
func FormatEuro(amount float64) string {
	// Format with 2 decimal places
	formatted := fmt.Sprintf("%.2f", amount)
	// Replace . with , for Euro format
	formatted = strings.Replace(formatted, ".", ",", 1)
	return "€" + formatted
}

// ParseEuro parses a Euro-formatted string to float64
// Example: "€100,50" -> 100.5
// Also handles formats without € symbol: "100,50" -> 100.5
func ParseEuro(euroStr string) (float64, error) {
	// Trim spaces first
	euroStr = strings.TrimSpace(euroStr)
	// Remove € symbol if present
	euroStr = strings.TrimPrefix(euroStr, "€")
	euroStr = strings.TrimSpace(euroStr)
	// Replace , with .
	euroStr = strings.Replace(euroStr, ",", ".", 1)
	// Parse float
	return strconv.ParseFloat(euroStr, 64)
}
