package handlers

import (
	"time"

	"github.com/shopspring/decimal"
)

// Helper functions used across handlers

// parseDate parses a date string in YYYY-MM-DD format
func parseDate(s string) (time.Time, error) {
	return time.Parse("2006-01-02", s)
}

// decimalFromFloat converts a float64 to decimal.Decimal
func decimalFromFloat(f float64) decimal.Decimal {
	return decimal.NewFromFloat(f)
}

// ptrTo returns a pointer to the given value
func ptrTo[T any](v T) *T {
	return &v
}
